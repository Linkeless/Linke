-- Rollback admin subscription management fields from user_subscriptions table

-- Drop indexes (safe to fail if not exists)
DROP INDEX IF EXISTS idx_resumed_by_admin_id ON user_subscriptions;
DROP INDEX IF EXISTS idx_paused_by_admin_id ON user_subscriptions;

-- Drop columns in reverse order of creation
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS max_pause_duration;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS resumed_by_admin_id;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS resumed_at;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS paused_by_admin_id;