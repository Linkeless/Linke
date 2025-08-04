-- Remove indexes for pause-related fields
DROP INDEX IF EXISTS idx_user_subscriptions_auto_resume;
DROP INDEX IF EXISTS idx_user_subscriptions_resumed_by_admin_id;
DROP INDEX IF EXISTS idx_user_subscriptions_resumed_at;
DROP INDEX IF EXISTS idx_user_subscriptions_paused_by_admin_id;
DROP INDEX IF EXISTS idx_user_subscriptions_paused_at;

-- Remove pause-related fields from user_subscriptions table
ALTER TABLE user_subscriptions
DROP COLUMN resumed_by_admin_id,
DROP COLUMN resumed_at,
DROP COLUMN max_pause_duration,
DROP COLUMN paused_by_admin_id,
DROP COLUMN pause_reason,
DROP COLUMN paused_at;