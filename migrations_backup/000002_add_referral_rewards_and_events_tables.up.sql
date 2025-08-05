-- ==============================================================================
-- ADD MISSING REFERRAL TABLES - UP MIGRATION
-- ==============================================================================
-- This migration adds the missing referral_rewards and referral_events tables
-- that were identified during the migration completeness audit.
-- ==============================================================================

-- ==============================================================================
-- REFERRAL REWARDS TABLE
-- ==============================================================================

-- Referral rewards for tracking and paying out referral bonuses
CREATE TABLE referral_rewards (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Foreign Keys
    referral_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    campaign_id BIGINT UNSIGNED NULL,
    
    -- Reward Details
    reward_type VARCHAR(50) NOT NULL,
    reward_amount DECIMAL(10,2) NOT NULL,
    reward_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    reward_description VARCHAR(255) NULL,
    
    -- Reward Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    earned_at TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    
    -- Payment Information
    payment_method VARCHAR(50) NULL,
    payment_reference VARCHAR(100) NULL,
    payment_data TEXT NULL,
    
    -- Conversion Details
    conversion_value DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    conversion_type VARCHAR(50) NULL,
    conversion_id BIGINT UNSIGNED NULL,
    
    -- Payout Information
    payout_batch_id BIGINT UNSIGNED NULL,
    payout_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    net_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Approval Workflow
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at TIMESTAMP NULL,
    approved_by_id BIGINT UNSIGNED NULL,
    rejected_at TIMESTAMP NULL,
    rejected_by_id BIGINT UNSIGNED NULL,
    rejection_reason VARCHAR(255) NULL,
    
    -- Metadata
    metadata TEXT NULL,
    notes TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_referral_rewards_referral (referral_id),
    INDEX idx_referral_rewards_user (user_id),
    INDEX idx_referral_rewards_campaign (campaign_id),
    INDEX idx_referral_rewards_reward_type (reward_type),
    INDEX idx_referral_rewards_status (status),
    INDEX idx_referral_rewards_earned_at (earned_at),
    INDEX idx_referral_rewards_paid_at (paid_at),
    INDEX idx_referral_rewards_expires_at (expires_at),
    INDEX idx_referral_rewards_payment_method (payment_method),
    INDEX idx_referral_rewards_payment_reference (payment_reference),
    INDEX idx_referral_rewards_conversion_type (conversion_type),
    INDEX idx_referral_rewards_conversion_id (conversion_id),
    INDEX idx_referral_rewards_payout_batch (payout_batch_id),
    INDEX idx_referral_rewards_approved_at (approved_at),
    INDEX idx_referral_rewards_approved_by (approved_by_id),
    INDEX idx_referral_rewards_rejected_at (rejected_at),
    INDEX idx_referral_rewards_rejected_by (rejected_by_id),
    INDEX idx_referral_rewards_created_at (created_at),
    INDEX idx_referral_rewards_deleted_at (deleted_at),
    
    -- Composite indexes for common queries
    INDEX idx_referral_rewards_user_status (user_id, status),
    INDEX idx_referral_rewards_referral_status (referral_id, status),
    INDEX idx_referral_rewards_status_earned (status, earned_at),
    INDEX idx_referral_rewards_status_paid (status, paid_at),
    INDEX idx_referral_rewards_approval_workflow (requires_approval, approved_at, status),
    INDEX idx_referral_rewards_payout_analysis (payout_batch_id, status, paid_at),
    
    -- Covering index for reward summaries
    INDEX idx_referral_rewards_covering (user_id, status, reward_type, reward_amount, reward_currency, earned_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Referral rewards tracking and payout management';

-- ==============================================================================
-- REFERRAL EVENTS TABLE
-- ==============================================================================

-- Referral events for tracking user actions and attribution
CREATE TABLE referral_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Foreign Keys
    referral_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Event Data
    event_type VARCHAR(50) NOT NULL,
    event_description VARCHAR(255) NULL,
    event_data TEXT NULL,
    
    -- Attribution
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    referrer_url VARCHAR(500) NULL,
    page_url VARCHAR(500) NULL,
    
    -- UTM Parameters
    utm_source VARCHAR(100) NULL,
    utm_campaign VARCHAR(100) NULL,
    utm_medium VARCHAR(100) NULL,
    utm_term VARCHAR(100) NULL,
    utm_content VARCHAR(100) NULL,
    
    -- Event Value
    event_value DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    event_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Metadata
    metadata TEXT NULL,
    processed_at TIMESTAMP NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_referral_events_referral (referral_id),
    INDEX idx_referral_events_user (user_id),
    INDEX idx_referral_events_event_type (event_type),
    INDEX idx_referral_events_ip_address (ip_address),
    INDEX idx_referral_events_utm_source (utm_source),
    INDEX idx_referral_events_utm_campaign (utm_campaign),
    INDEX idx_referral_events_utm_medium (utm_medium),
    INDEX idx_referral_events_processed_at (processed_at),
    INDEX idx_referral_events_created_at (created_at),
    INDEX idx_referral_events_deleted_at (deleted_at),
    
    -- Composite indexes for analytics
    INDEX idx_referral_events_referral_type_time (referral_id, event_type, created_at),
    INDEX idx_referral_events_user_type_time (user_id, event_type, created_at),
    INDEX idx_referral_events_type_time_value (event_type, created_at, event_value),
    INDEX idx_referral_events_utm_analysis (utm_source, utm_campaign, utm_medium, created_at),
    INDEX idx_referral_events_attribution (ip_address, user_agent, created_at),
    INDEX idx_referral_events_conversion_tracking (event_type, event_value, processed_at),
    
    -- Covering index for event summaries
    INDEX idx_referral_events_covering (referral_id, event_type, created_at, event_value, event_currency)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Referral event tracking for attribution and analytics';