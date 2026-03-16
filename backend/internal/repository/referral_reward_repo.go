package repository

import (
	"context"

	dbent "github.com/pengbin9472/ggbond/ent"
	"github.com/pengbin9472/ggbond/ent/referralreward"
	"github.com/pengbin9472/ggbond/ent/user"
	"github.com/pengbin9472/ggbond/internal/pkg/pagination"
	"github.com/pengbin9472/ggbond/internal/service"
)

type referralRewardRepository struct {
	client *dbent.Client
}

func NewReferralRewardRepository(client *dbent.Client) service.ReferralRewardRepository {
	return &referralRewardRepository{client: client}
}

func (r *referralRewardRepository) Create(ctx context.Context, reward *service.ReferralReward) error {
	client := clientFromContext(ctx, r.client)
	builder := client.ReferralReward.Create().
		SetInviterID(reward.InviterID).
		SetInviteeID(reward.InviteeID).
		SetRewardAmount(reward.RewardAmount).
		SetRewardType(reward.RewardType).
		SetTriggerCodeValue(reward.TriggerCodeValue).
		SetStatus(reward.Status)

	if reward.TriggerRedeemCodeID != nil {
		builder.SetTriggerRedeemCodeID(*reward.TriggerRedeemCodeID)
	}
	if reward.RewardRate != nil {
		builder.SetRewardRate(*reward.RewardRate)
	}
	if reward.Notes != "" {
		builder.SetNotes(reward.Notes)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	reward.ID = created.ID
	reward.CreatedAt = created.CreatedAt
	return nil
}

func (r *referralRewardRepository) ListByInviter(ctx context.Context, inviterID int64, params pagination.PaginationParams) ([]service.ReferralReward, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.ReferralReward.Query().
		Where(referralreward.InviterIDEQ(inviterID))

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	rewards, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(referralreward.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	inviteeEmails := make(map[int64]string, len(rewards))
	if len(rewards) > 0 {
		inviteeIDs := make([]int64, 0, len(rewards))
		seenInvitees := make(map[int64]struct{}, len(rewards))
		for _, reward := range rewards {
			if _, ok := seenInvitees[reward.InviteeID]; ok {
				continue
			}
			seenInvitees[reward.InviteeID] = struct{}{}
			inviteeIDs = append(inviteeIDs, reward.InviteeID)
		}

		users, err := client.User.Query().
			Where(user.IDIn(inviteeIDs...)).
			Select(user.FieldID, user.FieldEmail).
			All(ctx)
		if err != nil {
			return nil, nil, err
		}
		for _, u := range users {
			inviteeEmails[u.ID] = u.Email
		}
	}

	out := make([]service.ReferralReward, 0, len(rewards))
	for _, m := range rewards {
		item := referralRewardEntityToService(m)
		item.InviteeEmail = inviteeEmails[m.InviteeID]
		out = append(out, item)
	}

	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *referralRewardRepository) SumRewardsByInviter(ctx context.Context, inviterID int64) (float64, error) {
	var result []struct {
		Sum float64 `json:"sum"`
	}
	err := r.client.ReferralReward.Query().
		Where(referralreward.InviterIDEQ(inviterID)).
		Aggregate(dbent.As(dbent.Sum(referralreward.FieldRewardAmount), "sum")).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Sum, nil
}

func (r *referralRewardRepository) CountInviteesByInviter(ctx context.Context, inviterID int64) (int64, error) {
	// 从 users 表统计 referred_by = inviterID 的用户数量
	count, err := r.client.User.Query().
		Where(user.ReferredByEQ(inviterID)).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

func referralRewardEntityToService(m *dbent.ReferralReward) service.ReferralReward {
	out := service.ReferralReward{
		ID:               m.ID,
		InviterID:        m.InviterID,
		InviteeID:        m.InviteeID,
		RewardAmount:     m.RewardAmount,
		RewardType:       m.RewardType,
		RewardRate:       m.RewardRate,
		TriggerCodeValue: m.TriggerCodeValue,
		Status:           m.Status,
		CreatedAt:        m.CreatedAt,
	}
	if m.TriggerRedeemCodeID != nil {
		out.TriggerRedeemCodeID = m.TriggerRedeemCodeID
	}
	if m.Notes != nil {
		out.Notes = *m.Notes
	}
	return out
}
