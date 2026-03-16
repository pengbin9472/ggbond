package dto

import (
	"time"

	"github.com/pengbin9472/ggbond/internal/service"
)

// ReferralCodeDTO 邀请码响应
type ReferralCodeDTO struct {
	Code string `json:"code"`
}

// ReferralStatsDTO 邀请统计响应
type ReferralStatsDTO struct {
	TotalRewards float64 `json:"total_rewards"`
	InviteeCount int64   `json:"invitee_count"`
	RewardType   string  `json:"reward_type"`
	RewardRate   float64 `json:"reward_rate"`
}

// ReferralRewardDTO 返现记录响应
type ReferralRewardDTO struct {
	ID               int64     `json:"id"`
	InviterID        int64     `json:"inviter_id"`
	InviteeID        int64     `json:"invitee_id"`
	RewardAmount     float64   `json:"reward_amount"`
	RewardType       string    `json:"reward_type"`
	RewardRate       *float64  `json:"reward_rate"`
	TriggerCodeValue float64   `json:"trigger_code_value"`
	Status           string    `json:"status"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// ReferralStatsFromService 从 service 层转换统计数据
func ReferralStatsFromService(stats *service.ReferralStats) *ReferralStatsDTO {
	if stats == nil {
		return nil
	}
	return &ReferralStatsDTO{
		TotalRewards: stats.TotalRewards,
		InviteeCount: stats.InviteeCount,
		RewardType:   stats.RewardType,
		RewardRate:   stats.RewardRate,
	}
}

// ReferralRewardFromService 从 service 层转换返现记录
func ReferralRewardFromService(r *service.ReferralReward) *ReferralRewardDTO {
	if r == nil {
		return nil
	}
	return &ReferralRewardDTO{
		ID:               r.ID,
		InviterID:        r.InviterID,
		InviteeID:        r.InviteeID,
		RewardAmount:     r.RewardAmount,
		RewardType:       r.RewardType,
		RewardRate:       r.RewardRate,
		TriggerCodeValue: r.TriggerCodeValue,
		Status:           r.Status,
		Notes:            r.Notes,
		CreatedAt:        r.CreatedAt,
	}
}

// ReferralRewardsFromService 批量转换返现记录
func ReferralRewardsFromService(rewards []service.ReferralReward) []ReferralRewardDTO {
	out := make([]ReferralRewardDTO, 0, len(rewards))
	for i := range rewards {
		out = append(out, *ReferralRewardFromService(&rewards[i]))
	}
	return out
}
