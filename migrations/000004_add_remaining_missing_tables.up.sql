-- ==============================================================================
-- ADD REMAINING MISSING TABLES
-- ==============================================================================
-- This migration adds the 5 remaining tables that were missing from migration 3
-- The referrals table already exists, so we skip it and create only the missing ones
-- ==============================================================================

-- ==============================================================================
-- USAGE RECORDS TABLE (Critical - Usage tracking and billing)
-- ==============================================================================

CREATE TABLE usage_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Links to subscription (no FK constraint)
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    
    -- Usage Details
    usage_type VARCHAR(50) NOT NULL COMMENT 'Type of usage (traffic, api_calls, storage, etc.)',
    amount BIGINT NOT NULL COMMENT 'Usage amount in bytes or count',
    unit VARCHAR(20) NOT NULL DEFAULT 'bytes' COMMENT 'Unit of measurement',
    timestamp TIMESTAMP NOT NULL COMMENT 'When the usage occurred',
    
    -- Source Information
    source_type VARCHAR(50) NOT NULL COMMENT 'Source of usage (server, api, admin, etc.)',
    source_id VARCHAR(100) NULL COMMENT 'ID of the source',
    
    -- Additional Context
    metadata TEXT NULL COMMENT 'Additional metadata as JSON',
    user_agent TEXT NULL COMMENT 'User agent if applicable',
    ip_address VARCHAR(45) NULL COMMENT 'IP address if applicable',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_usage_records_subscription (user_subscription_id),
    INDEX idx_usage_records_usage_type (usage_type),
    INDEX idx_usage_records_timestamp (timestamp),
    INDEX idx_usage_records_source_type (source_type),
    INDEX idx_usage_records_created_at (created_at),
    INDEX idx_usage_records_deleted_at (deleted_at),
    
    -- Composite indexes for analytics
    INDEX idx_usage_records_subscription_type (user_subscription_id, usage_type),
    INDEX idx_usage_records_type_time (usage_type, timestamp),
    INDEX idx_usage_records_daily_summary (user_subscription_id, usage_type, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Detailed usage tracking records';

-- ==============================================================================
-- INVITE CODES TABLE (High Impact - Referral invitation system)
-- ==============================================================================

CREATE TABLE invite_codes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Core Fields
    code VARCHAR(32) NOT NULL,
    created_by_id BIGINT UNSIGNED NOT NULL,
    
    -- Referral Integration (no FK constraints)
    referral_campaign_id BIGINT UNSIGNED NULL,
    referral_reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    referral_reward_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Status and Limits
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    max_uses INT NOT NULL DEFAULT 10,
    used_count INT NOT NULL DEFAULT 0,
    
    -- Metadata
    description VARCHAR(255) NULL,
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraint on code
    UNIQUE INDEX idx_invite_codes_code (code),
    
    -- Essential indexes
    INDEX idx_invite_codes_created_by (created_by_id),
    INDEX idx_invite_codes_campaign (referral_campaign_id),
    INDEX idx_invite_codes_status (status),
    INDEX idx_invite_codes_created_at (created_at),
    INDEX idx_invite_codes_deleted_at (deleted_at),
    
    -- Composite indexes
    INDEX idx_invite_codes_creator_status (created_by_id, status),
    INDEX idx_invite_codes_usage_tracking (status, used_count, max_uses)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Invitation codes for referral system';

-- ==============================================================================
-- REFERRAL CAMPAIGNS TABLE (High Impact - Campaign management)
-- ==============================================================================

CREATE TABLE referral_campaigns (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Core Fields
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT NULL,
    
    -- Campaign Settings
    campaign_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timing
    start_date TIMESTAMP NULL,
    end_date TIMESTAMP NULL,
    
    -- Referrer Rewards
    referrer_reward_type VARCHAR(20) NOT NULL DEFAULT 'fixed',
    referrer_reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    referrer_reward_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    referrer_reward_cap DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Referee Rewards
    referee_reward_type VARCHAR(20) NOT NULL DEFAULT 'fixed',
    referee_reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    referee_reward_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Reward Conditions
    minimum_purchase_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    reward_trigger VARCHAR(50) NOT NULL DEFAULT 'registration',
    reward_delay INT NOT NULL DEFAULT 0,
    
    -- Limits
    max_referrals INT NOT NULL DEFAULT 0,
    max_rewards INT NOT NULL DEFAULT 0,
    total_reward_budget DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Targeting
    target_audience VARCHAR(100) NULL,
    eligible_user_segments TEXT NULL,
    restricted_countries TEXT NULL,
    
    -- Tracking
    tracking_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    conversion_goal VARCHAR(100) NULL,
    conversion_value DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Statistics
    total_referrals INT NOT NULL DEFAULT 0,
    total_conversions INT NOT NULL DEFAULT 0,
    total_rewards_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    conversion_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    
    -- Metadata
    metadata TEXT NULL,
    created_by_id BIGINT UNSIGNED NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraint on code
    UNIQUE INDEX idx_referral_campaigns_code (code),
    
    -- Essential indexes
    INDEX idx_referral_campaigns_name (name),
    INDEX idx_referral_campaigns_campaign_type (campaign_type),
    INDEX idx_referral_campaigns_status (status),
    INDEX idx_referral_campaigns_created_by (created_by_id),
    INDEX idx_referral_campaigns_created_at (created_at),
    INDEX idx_referral_campaigns_deleted_at (deleted_at),
    
    -- Composite indexes
    INDEX idx_referral_campaigns_active (status, start_date, end_date),
    INDEX idx_referral_campaigns_public_active (is_public, status),
    INDEX idx_referral_campaigns_performance (total_referrals, conversion_rate)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Referral marketing campaigns';

-- ==============================================================================
-- ALERT CONFIGURATIONS TABLE (Medium Impact - Usage alert setup)
-- ==============================================================================

CREATE TABLE alert_configurations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Links to subscription (no FK constraint)
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    
    -- Alert Settings
    usage_type VARCHAR(50) NOT NULL COMMENT 'Type of usage to monitor',
    threshold_type VARCHAR(20) NOT NULL DEFAULT 'percentage' COMMENT 'percentage or absolute',
    threshold DOUBLE NOT NULL COMMENT 'Alert threshold value',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT 'Whether alert is enabled',
    
    -- Notification Settings
    notification_channels TEXT NULL COMMENT 'JSON array of notification channels',
    cooldown_minutes INT NOT NULL DEFAULT 60 COMMENT 'Cooldown between alerts',
    
    -- Metadata
    name VARCHAR(100) NOT NULL COMMENT 'Human readable alert name',
    description TEXT NULL COMMENT 'Alert description',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT 'Alert priority',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_alert_configs_subscription (user_subscription_id),
    INDEX idx_alert_configs_usage_type (usage_type),
    INDEX idx_alert_configs_enabled (is_enabled),
    INDEX idx_alert_configs_priority (priority),
    INDEX idx_alert_configs_created_at (created_at),
    INDEX idx_alert_configs_deleted_at (deleted_at),
    
    -- Composite indexes
    INDEX idx_alert_configs_active (user_subscription_id, is_enabled),
    INDEX idx_alert_configs_monitoring (usage_type, is_enabled, threshold_type)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Alert configurations for usage monitoring';

-- ==============================================================================
-- USAGE ALERTS TABLE (Medium Impact - Usage alert management)
-- ==============================================================================

CREATE TABLE usage_alerts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Links (no FK constraints)
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    alert_configuration_id BIGINT UNSIGNED NOT NULL,
    
    -- Alert Details
    usage_type VARCHAR(50) NOT NULL,
    current_usage BIGINT NOT NULL COMMENT 'Usage when alert fired',
    usage_limit BIGINT NOT NULL COMMENT 'Usage limit at time of alert',
    threshold_value DOUBLE NOT NULL COMMENT 'Threshold that was exceeded',
    usage_percent DOUBLE NOT NULL COMMENT 'Percentage of limit used',
    
    -- Alert State
    status VARCHAR(20) NOT NULL DEFAULT 'fired' COMMENT 'Alert status',
    severity VARCHAR(20) NOT NULL COMMENT 'Alert severity level',
    fired_at TIMESTAMP NOT NULL COMMENT 'When alert was fired',
    resolved_at TIMESTAMP NULL COMMENT 'When alert was resolved',
    
    -- Notification Tracking
    notifications_sent INT NOT NULL DEFAULT 0 COMMENT 'Number of notifications sent',
    last_notification_sent TIMESTAMP NULL COMMENT 'Last notification time',
    notification_channels TEXT NULL COMMENT 'Channels used for notifications',
    notification_results TEXT NULL COMMENT 'Notification results as JSON',
    
    -- Additional Context
    message TEXT NULL COMMENT 'Alert message',
    metadata TEXT NULL COMMENT 'Additional metadata as JSON',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_usage_alerts_subscription (user_subscription_id),
    INDEX idx_usage_alerts_configuration (alert_configuration_id),
    INDEX idx_usage_alerts_usage_type (usage_type),
    INDEX idx_usage_alerts_status (status),
    INDEX idx_usage_alerts_severity (severity),
    INDEX idx_usage_alerts_fired_at (fired_at),
    INDEX idx_usage_alerts_resolved_at (resolved_at),
    INDEX idx_usage_alerts_created_at (created_at),
    INDEX idx_usage_alerts_deleted_at (deleted_at),
    
    -- Composite indexes
    INDEX idx_usage_alerts_active (user_subscription_id, status, resolved_at),
    INDEX idx_usage_alerts_monitoring (alert_configuration_id, status, fired_at),
    INDEX idx_usage_alerts_severity_status (severity, status, fired_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Fired usage alerts for monitoring and notification';