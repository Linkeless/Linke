-- Rollback remaining subscription fields

-- Drop indexes
DROP INDEX IF EXISTS idx_last_used_at ON user_subscriptions;
DROP INDEX IF EXISTS idx_last_renewal_failed ON user_subscriptions;

-- Drop columns in reverse order
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS notes;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS metadata;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS server_group_ids;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS usage_data;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS last_used_at;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS renewal_fail_reason;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS last_renewal_failed;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS renewal_attempts;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS auto_renew;