-- 用户表增加 referred_by 字段，记录邀请人
ALTER TABLE users ADD COLUMN IF NOT EXISTS referred_by BIGINT REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_users_referred_by ON users(referred_by);

-- 兑换码表增加 inviter_user_id 字段，标识邀请码归属用户
ALTER TABLE redeem_codes ADD COLUMN IF NOT EXISTS inviter_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_redeem_codes_inviter_user_id ON redeem_codes(inviter_user_id);

-- 返现记录表
CREATE TABLE IF NOT EXISTS referral_rewards (
    id BIGSERIAL PRIMARY KEY,
    inviter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trigger_redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    reward_amount DECIMAL(20,8) NOT NULL,
    reward_type VARCHAR(20) NOT NULL DEFAULT 'percentage',
    reward_rate DECIMAL(10,4),
    trigger_code_value DECIMAL(20,8) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_referral_rewards_inviter_id ON referral_rewards(inviter_id);
CREATE INDEX IF NOT EXISTS idx_referral_rewards_invitee_id ON referral_rewards(invitee_id);
CREATE INDEX IF NOT EXISTS idx_referral_rewards_created_at ON referral_rewards(created_at);
