package dto

import (
	"testing"
	"time"

	"github.com/pengbin9472/ggbond/internal/service"
	"github.com/stretchr/testify/require"
)

func TestReferralRewardFromServiceIncludesInviteeEmail(t *testing.T) {
	rewardRate := 20.0
	createdAt := time.Date(2026, 3, 17, 0, 0, 53, 0, time.UTC)
	reward := &service.ReferralReward{
		ID:               1,
		InviterID:        10,
		InviteeID:        20,
		InviteeEmail:     "invitee@example.com",
		RewardAmount:     20,
		RewardType:       "percentage",
		RewardRate:       &rewardRate,
		TriggerCodeValue: 100,
		Status:           "completed",
		CreatedAt:        createdAt,
	}

	result := ReferralRewardFromService(reward)

	require.NotNil(t, result)
	require.Equal(t, reward.InviteeEmail, result.InviteeEmail)
	require.Equal(t, reward.InviteeID, result.InviteeID)
	require.Equal(t, reward.CreatedAt, result.CreatedAt)
}
