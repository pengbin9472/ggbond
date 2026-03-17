package service

import (
	"context"
	"fmt"
	"strconv"

	dbent "github.com/pengbin9472/ggbond/ent"
	"github.com/pengbin9472/ggbond/internal/pkg/logger"
	"github.com/pengbin9472/ggbond/internal/pkg/pagination"
)

// ReferralRewardRepository 返现记录仓库接口
type ReferralRewardRepository interface {
	Create(ctx context.Context, reward *ReferralReward) error
	ListByInviter(ctx context.Context, inviterID int64, params pagination.PaginationParams) ([]ReferralReward, *pagination.PaginationResult, error)
	SumRewardsByInviter(ctx context.Context, inviterID int64) (float64, error)
	CountInviteesByInviter(ctx context.Context, inviterID int64) (int64, error)
}

// ReferralService 邀请返现服务
type ReferralService struct {
	rewardRepo           ReferralRewardRepository
	userRepo             UserRepository
	redeemRepo           RedeemCodeRepository
	settingService       *SettingService
	billingCacheService  *BillingCacheService
	authCacheInvalidator APIKeyAuthCacheInvalidator
	entClient            *dbent.Client
}

// NewReferralService 创建邀请返现服务
func NewReferralService(
	rewardRepo ReferralRewardRepository,
	userRepo UserRepository,
	redeemRepo RedeemCodeRepository,
	settingService *SettingService,
	billingCacheService *BillingCacheService,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	entClient *dbent.Client,
) *ReferralService {
	return &ReferralService{
		rewardRepo:           rewardRepo,
		userRepo:             userRepo,
		redeemRepo:           redeemRepo,
		settingService:       settingService,
		billingCacheService:  billingCacheService,
		authCacheInvalidator: authCacheInvalidator,
		entClient:            entClient,
	}
}

// GetOrCreateInvitationCode 获取或创建用户的专属邀请码
func (s *ReferralService) GetOrCreateInvitationCode(ctx context.Context, userID int64) (*RedeemCode, error) {
	// 查找用户已有的专属邀请码
	codes, _, err := s.redeemRepo.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 1}, RedeemTypeInvitation, StatusUnused, "")
	if err != nil {
		return nil, fmt.Errorf("list invitation codes: %w", err)
	}

	// 在结果中查找属于该用户的邀请码
	for i := range codes {
		if codes[i].InviterUserID != nil && *codes[i].InviterUserID == userID {
			return &codes[i], nil
		}
	}

	// 没有找到，需要全量搜索
	allCodes, _, err := s.redeemRepo.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 1000}, RedeemTypeInvitation, StatusUnused, "")
	if err != nil {
		return nil, fmt.Errorf("list all invitation codes: %w", err)
	}
	for i := range allCodes {
		if allCodes[i].InviterUserID != nil && *allCodes[i].InviterUserID == userID {
			return &allCodes[i], nil
		}
	}

	// 生成新的邀请码
	code, err := GenerateRedeemCode()
	if err != nil {
		return nil, fmt.Errorf("generate invitation code: %w", err)
	}

	redeemCode := &RedeemCode{
		Code:          code,
		Type:          RedeemTypeInvitation,
		Value:         0,
		Status:        StatusUnused,
		InviterUserID: &userID,
	}

	if err := s.redeemRepo.Create(ctx, redeemCode); err != nil {
		return nil, fmt.Errorf("create invitation code: %w", err)
	}

	return redeemCode, nil
}

// ProcessReferralReward 处理邀请返现
func (s *ReferralService) ProcessReferralReward(ctx context.Context, inviteeID int64, redeemCode *RedeemCode) error {
	// 检查返现功能是否启用
	if !s.isReferralEnabled(ctx) {
		return nil
	}

	// 仅余额类型触发返现
	if redeemCode.Type != RedeemTypeBalance {
		return nil
	}

	// 查询被邀请人的 referred_by
	invitee, err := s.userRepo.GetByID(ctx, inviteeID)
	if err != nil {
		return fmt.Errorf("get invitee: %w", err)
	}
	if invitee.ReferredBy == nil {
		return nil // 无邀请人，跳过
	}
	inviterID := *invitee.ReferredBy

	// 检查邀请人累计返现是否超限
	maxRewards := s.getMaxRewardsPerUser(ctx)
	if maxRewards > 0 {
		totalRewards, err := s.rewardRepo.SumRewardsByInviter(ctx, inviterID)
		if err != nil {
			return fmt.Errorf("sum rewards: %w", err)
		}
		if totalRewards >= maxRewards {
			logger.LegacyPrintf("service.referral", "[Referral] inviter %d reached max rewards limit %.2f", inviterID, maxRewards)
			return nil
		}
	}

	// 计算返现金额
	rewardAmount, rewardType, rewardRate := s.calculateReward(ctx, redeemCode.Value)
	if rewardAmount <= 0 {
		return nil
	}

	// 如果有上限，确保不超过
	if maxRewards > 0 {
		totalRewards, _ := s.rewardRepo.SumRewardsByInviter(ctx, inviterID)
		if totalRewards+rewardAmount > maxRewards {
			rewardAmount = maxRewards - totalRewards
			if rewardAmount <= 0 {
				return nil
			}
		}
	}

	// 在事务内：创建返现记录 + 给邀请人增加余额
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	// 创建返现记录
	reward := &ReferralReward{
		InviterID:           inviterID,
		InviteeID:           inviteeID,
		TriggerRedeemCodeID: &redeemCode.ID,
		RewardAmount:        rewardAmount,
		RewardType:          rewardType,
		RewardRate:          &rewardRate,
		TriggerCodeValue:    redeemCode.Value,
		Status:              "completed",
		Notes:               fmt.Sprintf("邀请用户 %d 兑换余额 %.2f，返现 %.2f", inviteeID, redeemCode.Value, rewardAmount),
	}
	if err := s.rewardRepo.Create(txCtx, reward); err != nil {
		return fmt.Errorf("create referral reward: %w", err)
	}

	// 给邀请人增加余额
	if err := s.userRepo.UpdateBalance(txCtx, inviterID, rewardAmount); err != nil {
		return fmt.Errorf("update inviter balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// 失效邀请人缓存
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, inviterID)
	}
	if s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateUserBalance(ctx, inviterID)
	}

	logger.LegacyPrintf("service.referral", "[Referral] inviter %d got reward %.4f from invitee %d redeem %.4f", inviterID, rewardAmount, inviteeID, redeemCode.Value)
	return nil
}

// GetInviterStats 获取邀请人统计
func (s *ReferralService) GetInviterStats(ctx context.Context, userID int64) (*ReferralStats, error) {
	totalRewards, err := s.rewardRepo.SumRewardsByInviter(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("sum rewards: %w", err)
	}

	inviteeCount, err := s.rewardRepo.CountInviteesByInviter(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count invitees: %w", err)
	}

	rewardType := s.getReferralRewardType(ctx)
	var rewardRate float64
	if rewardType == "fixed" {
		rewardRate = s.getFixedAmount(ctx)
	} else {
		rewardRate = s.getCashbackPercentage(ctx)
	}

	return &ReferralStats{
		TotalRewards: totalRewards,
		InviteeCount: inviteeCount,
		RewardType:   rewardType,
		RewardRate:   rewardRate,
	}, nil
}

// GetRewardHistory 获取返现历史
func (s *ReferralService) GetRewardHistory(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ReferralReward, *pagination.PaginationResult, error) {
	return s.rewardRepo.ListByInviter(ctx, userID, params)
}

// isReferralEnabled 检查返现功能是否启用
func (s *ReferralService) isReferralEnabled(ctx context.Context) bool {
	if s.settingService == nil {
		return false
	}
	value, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyReferralEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

// getReferralRewardType 获取返现类型
func (s *ReferralService) getReferralRewardType(ctx context.Context) string {
	if s.settingService == nil {
		return "percentage"
	}
	value, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyReferralRewardType)
	if err != nil || value == "" {
		return "percentage"
	}
	return value
}

// getCashbackPercentage 获取返现比例
func (s *ReferralService) getCashbackPercentage(ctx context.Context) float64 {
	if s.settingService == nil {
		return 10
	}
	value, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyReferralCashbackPercentage)
	if err != nil {
		return 10
	}
	if v, err := strconv.ParseFloat(value, 64); err == nil && v > 0 {
		return v
	}
	return 10
}

// getFixedAmount 获取固定返现金额
func (s *ReferralService) getFixedAmount(ctx context.Context) float64 {
	if s.settingService == nil {
		return 0
	}
	value, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyReferralFixedAmount)
	if err != nil {
		return 0
	}
	if v, err := strconv.ParseFloat(value, 64); err == nil && v > 0 {
		return v
	}
	return 0
}

// getMaxRewardsPerUser 获取每人最大返现总额
func (s *ReferralService) getMaxRewardsPerUser(ctx context.Context) float64 {
	if s.settingService == nil {
		return 0
	}
	value, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyReferralMaxRewardsPerUser)
	if err != nil {
		return 0
	}
	if v, err := strconv.ParseFloat(value, 64); err == nil && v > 0 {
		return v
	}
	return 0
}

// calculateReward 计算返现金额
func (s *ReferralService) calculateReward(ctx context.Context, codeValue float64) (amount float64, rewardType string, rate float64) {
	rewardType = s.getReferralRewardType(ctx)
	switch rewardType {
	case "fixed":
		rate = s.getFixedAmount(ctx)
		return rate, rewardType, rate
	default: // percentage
		rate = s.getCashbackPercentage(ctx)
		amount = codeValue * rate / 100
		return amount, "percentage", rate
	}
}
