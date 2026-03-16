package service

import "time"

// ReferralReward 邀请返现记录
type ReferralReward struct {
	ID                  int64
	InviterID           int64
	InviteeID           int64
	TriggerRedeemCodeID *int64
	RewardAmount        float64
	RewardType          string // "percentage" / "fixed"
	RewardRate          *float64
	TriggerCodeValue    float64
	Status              string
	Notes               string
	CreatedAt           time.Time
	Inviter             *User
	Invitee             *User
}

// ReferralStats 邀请统计
type ReferralStats struct {
	TotalRewards float64 `json:"total_rewards"`
	InviteeCount int64   `json:"invitee_count"`
	RewardType   string  `json:"reward_type"`
	RewardRate   float64 `json:"reward_rate"`
}
