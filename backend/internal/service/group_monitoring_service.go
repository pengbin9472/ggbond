package service

import (
	"context"
	"sync"
	"time"

	"github.com/pengbin9472/ggbond/internal/pkg/logger"
)

// GroupMonitoringService 分组监控服务
type GroupMonitoringService struct {
	groupRepo GroupRepository
	interval  time.Duration
	stopCh    chan struct{}
	cancelCtx context.CancelFunc
	wg        sync.WaitGroup
}

// NewGroupMonitoringService 创建分组监控服务
func NewGroupMonitoringService(groupRepo GroupRepository) *GroupMonitoringService {
	return &GroupMonitoringService{
		groupRepo: groupRepo,
		interval:  5 * time.Minute, // 默认每5分钟更新一次
		stopCh:    make(chan struct{}),
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

	// 为每次聚合设置 30 秒超时，避免 DB 慢查询阻塞退出
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 1. 从 accounts 表实时计算统计数据
	stats, err := s.groupRepo.ComputeGroupMonitoringStats(ctx)
	if err != nil {
		logger.LegacyPrintf("group_monitoring", "Failed to compute monitoring stats: %v", err)
		return
	}

	// 2. 更新统计表
	if err := s.groupRepo.UpsertGroupMonitoringStats(ctx, stats); err != nil {
		logger.LegacyPrintf("group_monitoring", "Failed to upsert monitoring stats: %v", err)
		return
	}

	// 3. 插入历史记录（用于趋势图表）
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
