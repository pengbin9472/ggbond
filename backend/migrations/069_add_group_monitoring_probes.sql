-- 069_add_group_monitoring_probes.sql
-- 为分组监控增加主动探针结果存储

CREATE TABLE IF NOT EXISTS group_monitoring_probes (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    account_id BIGINT NULL REFERENCES accounts(id) ON DELETE SET NULL,
    model TEXT NOT NULL DEFAULT '',
    success BOOLEAN NOT NULL DEFAULT FALSE,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    probed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_group_monitoring_probes_group_probed
    ON group_monitoring_probes(group_id, probed_at DESC);

CREATE INDEX IF NOT EXISTS idx_group_monitoring_probes_probed
    ON group_monitoring_probes(probed_at DESC);

ALTER TABLE group_monitoring_stats
    ADD COLUMN IF NOT EXISTS probe_status VARCHAR(16) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS last_probe_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_probe_success_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_probe_latency_ms INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_probe_error TEXT NOT NULL DEFAULT '';

COMMENT ON TABLE group_monitoring_probes IS '分组主动探针结果表 - 存储每次分组真实请求探测结果';
