package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pengbin9472/ggbond/internal/pkg/claude"
	"github.com/pengbin9472/ggbond/internal/pkg/ctxkey"
	"github.com/pengbin9472/ggbond/internal/pkg/geminicli"
	"github.com/pengbin9472/ggbond/internal/pkg/logger"
	"github.com/pengbin9472/ggbond/internal/pkg/openai"
)

// GroupMonitoringService 分组监控服务
type GroupMonitoringService struct {
	groupRepo       GroupRepository
	gatewaySvc      *GatewayService
	accountTestSvc  *AccountTestService
	interval        time.Duration
	probeTimeout    time.Duration
	maxProbeWorkers int
	stopCh          chan struct{}
	cancelCtx       context.CancelFunc
	wg              sync.WaitGroup
}

// NewGroupMonitoringService 创建分组监控服务
func NewGroupMonitoringService(groupRepo GroupRepository, gatewaySvc *GatewayService, accountTestSvc *AccountTestService) *GroupMonitoringService {
	return &GroupMonitoringService{
		groupRepo:       groupRepo,
		gatewaySvc:      gatewaySvc,
		accountTestSvc:  accountTestSvc,
		interval:        time.Minute,
		probeTimeout:    20 * time.Second,
		maxProbeWorkers: 4,
		stopCh:          make(chan struct{}),
	}
}

// Start 启动监控数据聚合任务
func (s *GroupMonitoringService) Start(ctx context.Context) {
	// 创建可取消的子 context
	ctx, cancel := context.WithCancel(ctx)
	s.cancelCtx = cancel

	s.wg.Add(1)
	go s.aggregationLoop(ctx)
	logger.LegacyPrintf("group_monitoring", "Group monitoring aggregation started (interval: %v)", s.interval)
}

// Stop 停止监控数据聚合任务
func (s *GroupMonitoringService) Stop() error {
	// 取消 context，中断进行中的 DB 操作
	if s.cancelCtx != nil {
		s.cancelCtx()
	}
	if s.stopCh != nil {
		close(s.stopCh)
	}
	s.wg.Wait()
	logger.LegacyPrintf("group_monitoring", "Group monitoring aggregation stopped")
	return nil
}

// aggregationLoop 聚合循环
func (s *GroupMonitoringService) aggregationLoop(ctx context.Context) {
	defer s.wg.Done()

	// 启动时立即执行一次
	s.aggregate(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.aggregate(ctx)
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// aggregate 执行一次聚合
func (s *GroupMonitoringService) aggregate(ctx context.Context) {
	start := time.Now()

	// 为每次聚合设置超时，避免探针或 DB 慢查询阻塞退出
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	// 1. 对活跃分组执行主动探针
	s.runGroupProbes(ctx)

	// 2. 从 accounts 表和探针结果计算统计数据
	stats, err := s.groupRepo.ComputeGroupMonitoringStats(ctx)
	if err != nil {
		logger.LegacyPrintf("group_monitoring", "Failed to compute monitoring stats: %v", err)
		return
	}

	// 3. 更新统计表
	if err := s.groupRepo.UpsertGroupMonitoringStats(ctx, stats); err != nil {
		logger.LegacyPrintf("group_monitoring", "Failed to upsert monitoring stats: %v", err)
		return
	}

	// 4. 插入历史记录（用于趋势图表）
	if err := s.groupRepo.InsertGroupMonitoringHistory(ctx, stats); err != nil {
		logger.LegacyPrintf("group_monitoring", "Failed to insert monitoring history: %v", err)
		return
	}

	duration := time.Since(start)
	logger.LegacyPrintf("group_monitoring", "Aggregation completed in %v, processed %d groups", duration, len(stats))
}

// RefreshNow 立即刷新监控数据
func (s *GroupMonitoringService) RefreshNow(ctx context.Context) error {
	s.aggregate(ctx)
	return nil
}

func (s *GroupMonitoringService) runGroupProbes(ctx context.Context) {
	if s == nil || s.groupRepo == nil || s.gatewaySvc == nil || s.accountTestSvc == nil {
		return
	}

	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		logger.LegacyPrintf("group_monitoring", "Failed to list active groups for probes: %v", err)
		return
	}
	if len(groups) == 0 {
		return
	}

	sem := make(chan struct{}, s.maxProbeWorkers)
	var wg sync.WaitGroup

	for i := range groups {
		group := groups[i]
		sem <- struct{}{}
		wg.Add(1)
		go func(g Group) {
			defer wg.Done()
			defer func() { <-sem }()
			s.probeOneGroup(ctx, g)
		}(group)
	}

	wg.Wait()
}

func (s *GroupMonitoringService) probeOneGroup(parent context.Context, group Group) {
	probeCtx, cancel := context.WithTimeout(parent, s.probeTimeout)
	defer cancel()

	probe := GroupMonitoringProbeResult{
		GroupID:  group.ID,
		Model:    defaultProbeModel(group.Platform),
		Success:  false,
		ProbedAt: time.Now(),
	}

	groupID := group.ID
	selectCtx := probeCtx
	if group.Platform == PlatformSora {
		selectCtx = context.WithValue(selectCtx, ctxkey.ForcePlatform, PlatformSora)
	}

	start := time.Now()
	account, err := s.gatewaySvc.SelectAccountForModel(selectCtx, &groupID, "", probe.Model)
	if err != nil {
		probe.LatencyMs = time.Since(start).Milliseconds()
		probe.ErrorMessage = fmt.Sprintf("select account failed: %v", err)
		if recErr := s.groupRepo.RecordGroupMonitoringProbe(parent, probe); recErr != nil {
			logger.LegacyPrintf("group_monitoring", "Failed to record failed probe for group=%d: %v", group.ID, recErr)
		}
		return
	}

	probe.AccountID = &account.ID
	result, err := s.accountTestSvc.RunTestBackground(probeCtx, account.ID, probe.Model)
	probe.LatencyMs = time.Since(start).Milliseconds()
	if result != nil && result.LatencyMs > 0 {
		probe.LatencyMs = result.LatencyMs
	}
	if err != nil {
		probe.ErrorMessage = err.Error()
	} else if result != nil {
		probe.Success = result.Status == "success"
		if !probe.Success {
			probe.ErrorMessage = result.ErrorMessage
		}
	}

	if recErr := s.groupRepo.RecordGroupMonitoringProbe(parent, probe); recErr != nil {
		logger.LegacyPrintf("group_monitoring", "Failed to record probe for group=%d account=%d: %v", group.ID, account.ID, recErr)
	}
}

func defaultProbeModel(platform string) string {
	switch platform {
	case PlatformOpenAI:
		return openai.DefaultTestModel
	case PlatformGemini:
		return geminicli.DefaultTestModel
	case PlatformAntigravity:
		return "claude-sonnet-4-5"
	case PlatformSora:
		return "sora2-landscape-10s"
	case PlatformAnthropic:
		fallthrough
	default:
		return claude.DefaultTestModel
	}
}
