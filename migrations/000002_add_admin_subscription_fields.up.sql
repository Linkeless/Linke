-- Add admin subscription management fields to user_subscriptions table
-- Required for AdminCreateUserSubscriptionRequest interface
-- Follows MySQL best practices: no foreign keys, conditional field addition

-- Add paused_by_admin_id field to track which admin paused subscription
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'paused_by_admin_id' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN paused_by_admin_id INT UNSIGNED NULL COMMENT \'Admin ID who paused subscription\'',
    'SELECT "paused_by_admin_id already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add resumed_at field to track when subscription was resumed
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'resumed_at' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN resumed_at TIMESTAMP NULL COMMENT \'When subscription was resumed\'',
    'SELECT "resumed_at already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add resumed_by_admin_id field to track which admin resumed subscription
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'resumed_by_admin_id' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN resumed_by_admin_id INT UNSIGNED NULL COMMENT \'Admin ID who resumed subscription\'',
    'SELECT "resumed_by_admin_id already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add max_pause_duration field with default 90 days
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'max_pause_duration' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN max_pause_duration INT NOT NULL DEFAULT 90 COMMENT \'Max pause duration in days\'',
    'SELECT "max_pause_duration already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add performance indexes (optional for MySQL best practices)
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND INDEX_NAME = 'idx_paused_by_admin_id' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@idx_exists = 0,
    'CREATE INDEX idx_paused_by_admin_id ON user_subscriptions(paused_by_admin_id)',
    'SELECT "idx_paused_by_admin_id already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND INDEX_NAME = 'idx_resumed_by_admin_id' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@idx_exists = 0,
    'CREATE INDEX idx_resumed_by_admin_id ON user_subscriptions(resumed_by_admin_id)',
    'SELECT "idx_resumed_by_admin_id already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;