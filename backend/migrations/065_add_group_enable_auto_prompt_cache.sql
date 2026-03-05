-- Migration 065: Add enable_auto_prompt_cache column to groups table
-- This column controls whether the gateway automatically injects cache_control
-- markers into Anthropic API requests for prompt caching.

ALTER TABLE groups ADD COLUMN IF NOT EXISTS enable_auto_prompt_cache BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN groups.enable_auto_prompt_cache IS '是否启用自动 Prompt 缓存（仅 Anthropic 平台）';
