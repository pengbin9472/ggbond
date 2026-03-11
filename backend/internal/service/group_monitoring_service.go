package service

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pengbin9472/ggbond/internal/pkg/claude"
	"github.com/pengbin9472/ggbond/internal/pkg/ctxkey"
	"github.com/pengbin9472/ggbond/internal/pkg/geminicli"
	"github.com/pengbin9472/ggbond/internal/pkg/logger"
)

// GroupMonitoringService 分组监控服务
type GroupMonitoringService struct {
	groupRepo       GroupRepository
	gatewaySvc      *GatewayService
	accountTestSvc  *AccountTestService
	interval        time.Duration
	probeTimeout    time.Duration
	maxProbeModels  int
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
		interval:        3 * time.Minute,
		probeTimeout:    12 * time.Second,
		maxProbeModels:  2,
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
		Success:  false,
		ProbedAt: time.Now(),
	}

	groupID := group.ID
	selectCtx := probeCtx
	if group.Platform == PlatformSora {
		selectCtx = context.WithValue(selectCtx, ctxkey.ForcePlatform, PlatformSora)
	}

	start := time.Now()
	probeModels := s.resolveGroupProbeModels(selectCtx, group)
	if s.maxProbeModels > 0 && len(probeModels) > s.maxProbeModels {
		probeModels = probeModels[:s.maxProbeModels]
	}
	var (
		account       *Account
		err           error
		selectedModel string
	)
	var lastProbeErr string
	for _, candidate := range probeModels {
		selectedModel = candidate
		account, err = s.gatewaySvc.SelectAccountForModel(selectCtx, &groupID, "", candidate)
		if err != nil {
			lastProbeErr = fmt.Sprintf("select account failed for model %s: %v", candidate, err)
			continue
		}

		probe.AccountID = &account.ID
		probe.Model = candidate
		result, testErr := s.accountTestSvc.RunTestBackground(probeCtx, account.ID, candidate)
		probe.LatencyMs = time.Since(start).Milliseconds()
		if result != nil && result.LatencyMs > 0 {
			probe.LatencyMs = result.LatencyMs
		}
		if testErr != nil {
			lastProbeErr = testErr.Error()
			continue
		}
		if result != nil && result.Status == "success" {
			probe.Success = true
			probe.ErrorMessage = ""
			break
		}
		if result != nil {
			lastProbeErr = result.ErrorMessage
		}
	}

	if !probe.Success {
		probe.LatencyMs = time.Since(start).Milliseconds()
		probe.Model = selectedModel
		if strings.TrimSpace(lastProbeErr) != "" {
			probe.ErrorMessage = lastProbeErr
		} else if err != nil {
			probe.ErrorMessage = err.Error()
		}
	}

	if recErr := s.groupRepo.RecordGroupMonitoringProbe(parent, probe); recErr != nil {
		if account != nil {
			logger.LegacyPrintf("group_monitoring", "Failed to record probe for group=%d account=%d: %v", group.ID, account.ID, recErr)
		} else {
			logger.LegacyPrintf("group_monitoring", "Failed to record probe for group=%d: %v", group.ID, recErr)
		}
	}
}

func (s *GroupMonitoringService) resolveGroupProbeModels(ctx context.Context, group Group) []string {
	candidates := make([]string, 0)
	seen := make(map[string]struct{})
	add := func(model string) {
		model = strings.TrimSpace(model)
		if model == "" {
			return
		}
		if _, exists := seen[model]; exists {
			return
		}
		seen[model] = struct{}{}
		candidates = append(candidates, model)
	}

	availableModels := []string(nil)
	if s.gatewaySvc != nil {
		availableModels = s.gatewaySvc.GetAvailableModels(ctx, &group.ID, group.Platform)
	}

	if len(availableModels) > 0 {
		exactModels := make([]string, 0, len(availableModels))
		for _, model := range availableModels {
			model = strings.TrimSpace(model)
			if model == "" || isWildcardModelPattern(model) {
				continue
			}
			exactModels = append(exactModels, model)
		}
		sort.SliceStable(exactModels, func(i, j int) bool {
			return compareProbeModelFreshness(exactModels[i], exactModels[j]) < 0
		})
		for _, model := range exactModels {
			add(model)
		}
	}

	for _, candidate := range preferredProbeModels(group.Platform) {
		if len(availableModels) == 0 || groupSupportsProbeModel(availableModels, candidate) {
			add(candidate)
		}
	}

	if len(candidates) == 0 {
		add(defaultProbeModel(group.Platform))
	}
	return candidates
}

func groupSupportsProbeModel(availableModels []string, candidate string) bool {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return false
	}
	for _, model := range availableModels {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if model == candidate {
			return true
		}
		if isWildcardModelPattern(model) && matchWildcard(model, candidate) {
			return true
		}
	}
	return false
}

func isWildcardModelPattern(model string) bool {
	return strings.Contains(model, "*")
}

var probeModelNumberPattern = regexp.MustCompile(`\d+`)

func compareProbeModelFreshness(left, right string) int {
	leftDate := extractProbeModelDate(left)
	rightDate := extractProbeModelDate(right)
	if !leftDate.Equal(rightDate) {
		if leftDate.After(rightDate) {
			return -1
		}
		return 1
	}

	leftNums := extractProbeModelNumbers(left)
	rightNums := extractProbeModelNumbers(right)
	for i := 0; i < len(leftNums) && i < len(rightNums); i++ {
		if leftNums[i] == rightNums[i] {
			continue
		}
		if leftNums[i] > rightNums[i] {
			return -1
		}
		return 1
	}
	if len(leftNums) != len(rightNums) {
		if len(leftNums) > len(rightNums) {
			return -1
		}
		return 1
	}
	return strings.Compare(left, right)
}

func extractProbeModelDate(model string) time.Time {
	parts := probeModelNumberPattern.FindAllString(model, -1)
	for i := 0; i+2 < len(parts); i++ {
		if len(parts[i]) == 4 && len(parts[i+1]) == 2 && len(parts[i+2]) == 2 {
			year, _ := strconv.Atoi(parts[i])
			month, _ := strconv.Atoi(parts[i+1])
			day, _ := strconv.Atoi(parts[i+2])
			if year >= 2000 && month >= 1 && month <= 12 && day >= 1 && day <= 31 {
				return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			}
		}
	}
	return time.Time{}
}

func extractProbeModelNumbers(model string) []int {
	matches := probeModelNumberPattern.FindAllString(model, -1)
	out := make([]int, 0, len(matches))
	for _, match := range matches {
		value, err := strconv.Atoi(match)
		if err != nil {
			continue
		}
		out = append(out, value)
	}
	return out
}

func preferredProbeModels(platform string) []string {
	switch platform {
	case PlatformOpenAI:
		return []string{"gpt-4o-mini", "gpt-4.1-mini", "gpt-4.1", "gpt-5.1", "gpt-5.1-codex"}
	case PlatformGemini:
		return []string{geminicli.DefaultTestModel, "gemini-2.5-flash", "gemini-2.5-pro"}
	case PlatformAntigravity:
		return []string{"claude-sonnet-4-5", "gemini-3-flash", "claude-haiku-4-5"}
	case PlatformSora:
		return []string{"sora2-landscape-10s", "prompt-enhance-short-10s"}
	case PlatformAnthropic:
		fallthrough
	default:
		return []string{claude.DefaultTestModel, "claude-sonnet-4-5", "claude-haiku-4-5"}
	}
}

func defaultProbeModel(platform string) string {
	switch platform {
	case PlatformOpenAI:
		return "gpt-4o-mini"
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
