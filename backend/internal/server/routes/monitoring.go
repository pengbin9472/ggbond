package routes

import (
	"github.com/pengbin9472/ggbond/internal/handler"
	"github.com/pengbin9472/ggbond/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterMonitoringRoutes 注册监控相关路由（普通用户和管理员都可访问）
func RegisterMonitoringRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
) {
	// 需要登录才能访问
	authenticated := v1.Group("/monitoring")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	{
		// 分组监控统计
		authenticated.GET("/groups", h.Monitoring.GetGroupMonitoring)
		// 分组监控历史
		authenticated.GET("/groups/:id/history", h.Monitoring.GetGroupMonitoringHistory)
	}
}
