-- Add remaining fields that may be missing from user_subscriptions table
-- This ensures full compatibility with UserSubscription entity

-- Auto-renewal fields
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'auto_renew' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN auto_renew BOOLEAN NOT NULL DEFAULT FALSE COMMENT \'Whether auto renewal is enabled\'',
    'SELECT "auto_renew already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'renewal_attempts' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN renewal_attempts INT NOT NULL DEFAULT 0 COMMENT \'Number of renewal attempts\'',
    'SELECT "renewal_attempts already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'last_renewal_failed' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN last_renewal_failed TIMESTAMP NULL COMMENT \'Last renewal failure time\'',
    'SELECT "last_renewal_failed already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'renewal_fail_reason' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN renewal_fail_reason VARCHAR(255) COMMENT \'Reason for renewal failure\'',
    'SELECT "renewal_fail_reason already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Usage tracking fields
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'last_used_at' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN last_used_at TIMESTAMP NULL COMMENT \'Last time subscription was used\'',
    'SELECT "last_used_at already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'usage_data' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN usage_data TEXT COMMENT \'Usage data in JSON format\'',
    'SELECT "usage_data already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Server group access
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'server_group_ids' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN server_group_ids TEXT COMMENT \'Accessible server group IDs in JSON format\'',
    'SELECT "server_group_ids already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Metadata and notes
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'metadata' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN metadata TEXT COMMENT \'Additional metadata in JSON format\'',
    'SELECT "metadata already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND COLUMN_NAME = 'notes' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@col_exists = 0,
    'ALTER TABLE user_subscriptions ADD COLUMN notes TEXT COMMENT \'Admin notes and comments\'',
    'SELECT "notes already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add performance indexes for renewal tracking
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND INDEX_NAME = 'idx_last_renewal_failed' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@idx_exists = 0,
    'CREATE INDEX idx_last_renewal_failed ON user_subscriptions(last_renewal_failed)',
    'SELECT "idx_last_renewal_failed already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_NAME = 'user_subscriptions' 
    AND INDEX_NAME = 'idx_last_used_at' 
    AND TABLE_SCHEMA = DATABASE());

SET @sql = IF(@idx_exists = 0,
    'CREATE INDEX idx_last_used_at ON user_subscriptions(last_used_at)',
    'SELECT "idx_last_used_at already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;