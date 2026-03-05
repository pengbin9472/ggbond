package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pengbin9472/ggbond/internal/pkg/response"
	"github.com/pengbin9472/ggbond/internal/server/middleware"
	"github.com/pengbin9472/ggbond/internal/service"
)

// MonitoringHandler 监控相关接口处理器
type MonitoringHandler struct {
	groupRepo   service.GroupRepository
	userRepo    service.UserRepository
	userSubRepo service.UserSubscriptionRepository
}

// NewMonitoringHandler 创建监控处理器
func NewMonitoringHandler(groupRepo service.GroupRepository, userRepo service.UserRepository, userSubRepo service.UserSubscriptionRepository) *MonitoringHandler {
	return &MonitoringHandler{
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		userSubRepo: userSubRepo,
	}
}

// GetGroupMonitoring 获取分组监控统计
// GET /api/v1/monitoring/groups
func (h *MonitoringHandler) GetGroupMonitoring(c *gin.Context) {
	stats, err := h.groupRepo.GetGroupMonitoringStats(c.Request.Context())
	if err != nil {
		response.Error(c, 500, "Failed to get group monitoring stats")
		return
	}

	// 获取当前用户信息，按权限过滤分组
	filtered := stats
	authSubject, ok := middleware.GetAuthSubjectFromContext(c)
	if ok {
		user, err := h.userRepo.GetByID(c.Request.Context(), authSubject.UserID)
		if err == nil && !user.IsAdmin() {
			// 非管理员用户：只显示有权限的分组
			// 获取用户的有效订阅
			subscribedGroupIDs := make(map[int64]bool)
			activeSubscriptions, subErr := h.userSubRepo.ListActiveByUserID(c.Request.Context(), user.ID)
			if subErr == nil {
				for _, sub := range activeSubscriptions {
					subscribedGroupIDs[sub.GroupID] = true
				}
			}

			filtered = make([]service.GroupMonitoringStat, 0, len(stats))
			for _, stat := range stats {
				// 订阅类型分组：需要有效订阅
				if stat.SubscriptionType == service.SubscriptionTypeSubscription {
					if subscribedGroupIDs[stat.GroupID] {
						filtered = append(filtered, stat)
					}
					continue
				}
				// 标准类型分组：使用 CanBindGroup 逻辑
				if user.CanBindGroup(stat.GroupID, stat.IsExclusive) {
					filtered = append(filtered, stat)
				}
			}
		}
	}

	response.Success(c, gin.H{
		"groups": filtered,
	})
}

// GetGroupMonitoringHistory 获取分组监控历史数据
// GET /api/v1/monitoring/groups/:id/history
func (h *MonitoringHandler) GetGroupMonitoringHistory(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, 400, "Invalid group ID")
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	history, err := h.groupRepo.GetGroupMonitoringHistory(c.Request.Context(), groupID, limit)
	if err != nil {
		response.Error(c, 500, "Failed to get monitoring history")
		return
	}

	response.Success(c, gin.H{
		"history": history,
	})
}
