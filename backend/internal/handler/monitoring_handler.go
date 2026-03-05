package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pengbin9472/ggbond/internal/pkg/response"
	"github.com/pengbin9472/ggbond/internal/service"
)

// MonitoringHandler 监控相关接口处理器
type MonitoringHandler struct {
	groupRepo service.GroupRepository
}

// NewMonitoringHandler 创建监控处理器
func NewMonitoringHandler(groupRepo service.GroupRepository) *MonitoringHandler {
	return &MonitoringHandler{
		groupRepo: groupRepo,
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

	response.Success(c, gin.H{
		"groups": stats,
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
