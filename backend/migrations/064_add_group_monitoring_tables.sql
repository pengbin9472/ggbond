-- 064_add_group_monitoring_tables.sql
-- 添加分组监控相关表

-- 分组监控统计表（存储每个分组的实时统计快照）
CREATE TABLE IF NOT EXISTS group_monitoring_stats (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,

    -- 账户状态统计
    total_accounts INTEGER NOT NULL DEFAULT 0,
    normal_accounts INTEGER NOT NULL DEFAULT 0,
    error_accounts INTEGER NOT NULL DEFAULT 0,
    ratelimit_accounts INTEGER NOT NULL DEFAULT 0,
    overload_accounts INTEGER NOT NULL DEFAULT 0,
    disabled_accounts INTEGER NOT NULL DEFAULT 0,

    -- 性能指标（从 usage_logs 聚合）
    avg_response_time INTEGER DEFAULT 0,
    availability_rate DECIMAL(8,4) DEFAULT -1,
    cache_hit_rate DECIMAL(8,4) DEFAULT -1,

    -- 时间戳
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 唯一约束：每个分组只有一条记录
    CONSTRAINT uk_group_monitoring_stats_group_id UNIQUE (group_id)
);

CREATE INDEX IF NOT EXISTS idx_group_monitoring_stats_updated_at ON group_monitoring_stats(updated_at);

-- 分组监控历史表（存储趋势数据，按分钟去重防止多实例重复写入）
CREATE TABLE IF NOT EXISTS group_monitoring_history (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,

    -- 账户状态快照
    total_accounts INTEGER NOT NULL DEFAULT 0,
    normal_accounts INTEGER NOT NULL DEFAULT 0,
    error_accounts INTEGER NOT NULL DEFAULT 0,
    ratelimit_accounts INTEGER NOT NULL DEFAULT 0,
    overload_accounts INTEGER NOT NULL DEFAULT 0,

    -- 性能指标
    availability_rate DECIMAL(8,4) DEFAULT -1,
    cache_hit_rate DECIMAL(8,4) DEFAULT -1,

    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 按分钟截断的唯一索引：同一分组同一分钟只允许一条记录，多实例 upsert 安全
-- 注意：date_trunc 默认是 STABLE 不是 IMMUTABLE，需要创建 IMMUTABLE 包装函数
CREATE OR REPLACE FUNCTION date_trunc_immutable(text, timestamptz)
RETURNS timestamptz AS $$
    SELECT date_trunc($1, $2);
$$ LANGUAGE SQL IMMUTABLE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_monitoring_history_group_minute
    ON group_monitoring_history(group_id, date_trunc_immutable('minute', recorded_at));

CREATE INDEX IF NOT EXISTS idx_group_monitoring_history_group_recorded
    ON group_monitoring_history(group_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_group_monitoring_history_recorded
    ON group_monitoring_history(recorded_at DESC);

COMMENT ON TABLE group_monitoring_stats IS '分组监控统计表 - 存储每个分组的实时账户和请求统计';
COMMENT ON TABLE group_monitoring_history IS '分组监控历史表 - 存储可用率和缓存命中率趋势数据';
