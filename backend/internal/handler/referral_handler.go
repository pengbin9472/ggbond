package handler

import (
	"strconv"

	"github.com/pengbin9472/ggbond/internal/handler/dto"
	"github.com/pengbin9472/ggbond/internal/pkg/pagination"
	"github.com/pengbin9472/ggbond/internal/pkg/response"
	middleware2 "github.com/pengbin9472/ggbond/internal/server/middleware"
	"github.com/pengbin9472/ggbond/internal/service"

	"github.com/gin-gonic/gin"
)

// ReferralHandler handles referral-related requests
type ReferralHandler struct {
	referralService *service.ReferralService
}

// NewReferralHandler creates a new ReferralHandler
func NewReferralHandler(referralService *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{
		referralService: referralService,
	}
}

// GetInvitationCode 获取/生成我的专属邀请码
// GET /api/v1/referrals/code
func (h *ReferralHandler) GetInvitationCode(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	code, err := h.referralService.GetOrCreateInvitationCode(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.ReferralCodeDTO{Code: code.Code})
}

// GetStats 获取我的邀请统计
// GET /api/v1/referrals/stats
func (h *ReferralHandler) GetStats(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	stats, err := h.referralService.GetInviterStats(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.ReferralStatsFromService(stats))
}

// GetHistory 获取返现记录
// GET /api/v1/referrals/history
func (h *ReferralHandler) GetHistory(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	rewards, pag, err := h.referralService.GetRewardHistory(c.Request.Context(), subject.UserID, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":      dto.ReferralRewardsFromService(rewards),
		"pagination": pag,
	})
}
