-- Add pause-related fields to user_subscriptions table
ALTER TABLE user_subscriptions
ADD COLUMN paused_at TIMESTAMP NULL,
ADD COLUMN pause_reason VARCHAR(255) DEFAULT '',
ADD COLUMN paused_by_admin_id INTEGER NULL,
ADD COLUMN max_pause_duration INTEGER NOT NULL DEFAULT 90 COMMENT 'Maximum pause duration in days',
ADD COLUMN resumed_at TIMESTAMP NULL,
ADD COLUMN resumed_by_admin_id INTEGER NULL;

-- Add indexes for pause-related fields for efficient querying
CREATE INDEX idx_user_subscriptions_paused_at ON user_subscriptions(paused_at);
CREATE INDEX idx_user_subscriptions_paused_by_admin_id ON user_subscriptions(paused_by_admin_id);
CREATE INDEX idx_user_subscriptions_resumed_at ON user_subscriptions(resumed_at);
CREATE INDEX idx_user_subscriptions_resumed_by_admin_id ON user_subscriptions(resumed_by_admin_id);

-- Add index for finding subscriptions that should be auto-resumed
CREATE INDEX idx_user_subscriptions_auto_resume ON user_subscriptions(status, paused_at, max_pause_duration);