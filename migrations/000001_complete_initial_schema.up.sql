-- Complete Initial Database Schema
-- This migration includes the complete schema with all features:
-- - User management with OAuth support
-- - Subscription system with billing
-- - Payment processing
-- - Referral system
-- - Support tickets
-- - Server management
-- - Traffic management system
-- - Performance optimizations
--
-- IMPORTANT: Foreign key constraints are handled at application level
-- for better performance and flexibility. No database-level FK constraints.

-- ==============================================================================
-- USERS TABLE - WITH ALL OPTIMIZATIONS
-- ==============================================================================

CREATE TABLE users (
    -- Primary key
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Authentication core fields
    email VARCHAR(191) NOT NULL,
    password VARCHAR(255) NULL,
    
    -- Status as string for readability
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    provider VARCHAR(20) NOT NULL DEFAULT 'local',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Profile fields
    username VARCHAR(32) NULL,
    name VARCHAR(128) NULL,
    phone VARCHAR(20) NULL,
    avatar VARCHAR(255) NULL,
    
    -- OAuth provider IDs
    google_id VARCHAR(32) NULL,
    github_id VARCHAR(32) NULL,
    telegram_id VARCHAR(32) NULL,
    
    -- Security fields
    failed_login_attempts TINYINT UNSIGNED NOT NULL DEFAULT 0,
    login_count INT UNSIGNED NOT NULL DEFAULT 0,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45) NULL,
    locked_until TIMESTAMP NULL,
    
    -- Verification timestamps
    email_verified_at TIMESTAMP NULL,
    phone_verified_at TIMESTAMP NULL,
    
    -- Referral fields
    total_referrals SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    total_referral_rewards DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    invite_code_id BIGINT UNSIGNED NULL,
    invite_code_used VARCHAR(32) NULL,
    referral_code VARCHAR(16) NULL,
    referral_link VARCHAR(255) NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- OAuth provider data (JSON)
    provider_data JSON NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes only
    UNIQUE KEY uk_users_email (email),
    UNIQUE KEY uk_users_google_id (google_id),
    UNIQUE KEY uk_users_github_id (github_id),
    UNIQUE KEY uk_users_telegram_id (telegram_id),
    
    -- Performance indexes
    INDEX idx_users_status (status),
    INDEX idx_users_provider (provider, google_id, github_id, telegram_id),
    INDEX idx_users_referral (referral_code),
    INDEX idx_users_created (created_at),
    INDEX idx_users_deleted (deleted_at),
    
    -- Performance optimization covering index
    INDEX idx_users_covering (id, email, username, status)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- INVITE CODES TABLE
-- ==============================================================================

CREATE TABLE invite_codes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code CHAR(32) NOT NULL,
    created_by_id BIGINT UNSIGNED NOT NULL,
    
    -- Status as string for readability
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Usage tracking
    max_uses SMALLINT UNSIGNED NOT NULL DEFAULT 10,
    used_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    last_used_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    
    -- Metadata
    description VARCHAR(128) NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes only
    UNIQUE KEY uk_invite_codes_code (code),
    INDEX idx_invite_codes_creator (created_by_id),
    INDEX idx_invite_codes_active (is_active, status, expires_at),
    INDEX idx_invite_codes_deleted (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Invite code usage audit trail
CREATE TABLE invite_code_usages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    invite_code_id BIGINT UNSIGNED NOT NULL,
    used_by_id BIGINT UNSIGNED NOT NULL,
    used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(255) NULL,
    
    PRIMARY KEY (id),
    INDEX idx_usage_invite_code (invite_code_id),
    INDEX idx_usage_user (used_by_id),
    INDEX idx_usage_time (used_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- SUBSCRIPTION PLANS TABLE
-- ==============================================================================

CREATE TABLE subscription_plans (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT NULL,
    
    -- Pricing
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_cycle VARCHAR(20) NOT NULL,
    billing_interval INT NOT NULL DEFAULT 1,
    
    -- Features
    features TEXT NULL,
    limits TEXT NULL,
    
    -- Display settings
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    is_popular BOOLEAN NOT NULL DEFAULT FALSE,
    is_recommended BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Fees
    setup_fee DECIMAL(10,2) NOT NULL DEFAULT 0,
    cancellation_fee DECIMAL(10,2) NOT NULL DEFAULT 0,
    trial_period_days INT NOT NULL DEFAULT 0,
    
    -- Traffic Configuration (Required for all plans)
    traffic_limit BIGINT NOT NULL DEFAULT 10737418240 COMMENT 'Traffic limit in bytes (default: 10GB, 0 = unlimited)',
    traffic_reset_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly' COMMENT 'Traffic reset cycle',
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    UNIQUE KEY uk_plans_code (code),
    INDEX idx_plans_status (status, is_visible),
    INDEX idx_plans_currency (currency),
    INDEX idx_plans_deleted (deleted_at),
    INDEX idx_plans_is_recommended (is_recommended),
    
    -- Performance optimization covering index
    INDEX idx_subscription_plans_covering (id, name, price, status)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- USER SUBSCRIPTIONS TABLE - WITH ALL ENHANCEMENTS
-- ==============================================================================

CREATE TABLE user_subscriptions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    uuid VARCHAR(36) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_plan_id BIGINT UNSIGNED NOT NULL,
    
    -- Status and billing
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    is_trial BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Pricing
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_cycle VARCHAR(20) NOT NULL,
    billing_interval INT NOT NULL DEFAULT 1,
    
    -- Important dates
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NULL,
    current_period_start TIMESTAMP NULL,
    current_period_end TIMESTAMP NULL,
    next_billing_date TIMESTAMP NULL,
    trial_end_date TIMESTAMP NULL,
    
    -- Cancellation
    cancelled_at TIMESTAMP NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    cancellation_reason VARCHAR(255) NULL,
    
    -- Payment tracking
    payment_failed_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
    last_payment_attempt TIMESTAMP NULL,
    last_used_at TIMESTAMP NULL,
    
    -- Usage data
    usage_data TEXT NULL,
    notes TEXT NULL,
    metadata TEXT NULL,
    
    -- Server Group Access
    server_group_ids TEXT DEFAULT NULL COMMENT 'JSON array of server group IDs that this subscription can access',
    
    -- Traffic Configuration and Usage
    traffic_limit BIGINT NOT NULL DEFAULT 0 COMMENT 'Total traffic limit in bytes (0 = unlimited)',
    traffic_used BIGINT NOT NULL DEFAULT 0 COMMENT 'Total traffic used in bytes',
    traffic_reset_date TIMESTAMP NULL COMMENT 'Next traffic reset date',
    traffic_reset_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly' COMMENT 'Traffic reset cycle',
    traffic_suspended BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Whether account is suspended due to traffic limit',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    UNIQUE KEY uk_user_subs_uuid (uuid),
    UNIQUE KEY uk_user_subscriptions_uuid (uuid),
    INDEX idx_user_subs_user (user_id, status),
    INDEX idx_user_subs_billing (next_billing_date, is_auto_renew),
    INDEX idx_user_subs_active (is_active, status),
    INDEX idx_user_subs_plan (subscription_plan_id),
    INDEX idx_user_subs_deleted (deleted_at),
    
    -- Performance optimization indexes
    INDEX idx_user_subs_user_status_created (user_id, status, created_at DESC),
    INDEX idx_user_subs_plan_status_created (subscription_plan_id, status, created_at DESC),
    
    -- Traffic management indexes
    INDEX idx_user_subscriptions_traffic_limit (traffic_limit),
    INDEX idx_user_subscriptions_traffic_used (traffic_used),
    INDEX idx_user_subscriptions_traffic_reset_date (traffic_reset_date),
    INDEX idx_user_subscriptions_traffic_suspended (traffic_suspended),
    INDEX idx_user_subscriptions_traffic_status (traffic_suspended, status, deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- TRAFFIC MANAGEMENT TABLES
-- ==============================================================================

-- Node data table for storing UniProxy push data
CREATE TABLE node_data (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    node_id INT UNSIGNED NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    token VARCHAR(255) NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    content_type VARCHAR(100),
    user_agent VARCHAR(255),
    request_body LONGTEXT NOT NULL,
    processed_at TIMESTAMP NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes for performance
    INDEX idx_node_data_node_id (node_id),
    INDEX idx_node_data_node_type (node_type),
    INDEX idx_node_data_token (token),
    INDEX idx_node_data_processed_at (processed_at),
    INDEX idx_node_data_status (status),
    INDEX idx_node_data_created_at (created_at),
    INDEX idx_node_data_deleted_at (deleted_at)
);

-- User traffic logs table for storing traffic usage data
CREATE TABLE user_traffic_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    node_id INT UNSIGNED NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    upload_bytes BIGINT NOT NULL DEFAULT 0,
    download_bytes BIGINT NOT NULL DEFAULT 0,
    total_bytes BIGINT NOT NULL DEFAULT 0,
    node_data_id BIGINT UNSIGNED NOT NULL,
    recorded_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes for performance
    INDEX idx_user_traffic_logs_user_id (user_id),
    INDEX idx_user_traffic_logs_node_id (node_id),
    INDEX idx_user_traffic_logs_node_type (node_type),
    INDEX idx_user_traffic_logs_total_bytes (total_bytes),
    INDEX idx_user_traffic_logs_node_data_id (node_data_id),
    INDEX idx_user_traffic_logs_recorded_at (recorded_at),
    INDEX idx_user_traffic_logs_created_at (created_at),
    INDEX idx_user_traffic_logs_deleted_at (deleted_at),
    
    -- Composite indexes for common queries
    INDEX idx_user_traffic_logs_user_node (user_id, node_id, recorded_at),
    INDEX idx_user_traffic_logs_node_time (node_id, recorded_at)
    
    -- Note: Foreign key constraints are handled at application level, not database level
);

-- ==============================================================================
-- PAYMENT RECORDS TABLE
-- ==============================================================================

CREATE TABLE payment_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    payment_no CHAR(32) NOT NULL,
    out_trade_no CHAR(32) NOT NULL,
    transaction_id VARCHAR(64) NULL,
    
    -- User and order references
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_order_id BIGINT UNSIGNED NULL,
    
    -- Payment details
    gateway VARCHAR(30) NOT NULL,
    payment_method VARCHAR(30) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'CNY',
    
    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(20) NULL,
    
    -- Important timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    paid_at TIMESTAMP NULL,
    expired_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    
    -- Refund info
    refund_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    refund_status VARCHAR(20) NULL,
    refunded_at TIMESTAMP NULL,
    refund_reason VARCHAR(100) NULL,
    
    -- Retry tracking
    retry_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
    last_retry_at TIMESTAMP NULL,
    
    -- Client info
    client_ip VARCHAR(45) NULL,
    user_agent TEXT NULL,
    
    -- URLs and responses
    payment_url VARCHAR(512) NULL,
    notify_url VARCHAR(512) NULL,
    return_url VARCHAR(512) NULL,
    gateway_response JSON NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    UNIQUE KEY uk_payment_no (payment_no),
    UNIQUE KEY uk_out_trade_no (out_trade_no),
    INDEX idx_payment_user (user_id, status),
    INDEX idx_payment_gateway (gateway, status),
    INDEX idx_payment_status (status, paid_at),
    INDEX idx_payment_order (subscription_order_id),
    INDEX idx_payment_created (created_at),
    INDEX idx_payment_deleted (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- SUBSCRIPTION ORDERS TABLE
-- ==============================================================================

CREATE TABLE subscription_orders (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_number VARCHAR(50) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_plan_id BIGINT UNSIGNED NOT NULL,
    user_subscription_id BIGINT UNSIGNED NULL,
    
    -- Order details
    order_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- Financial details
    amount DECIMAL(10,2) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    
    -- Payment info
    payment_method VARCHAR(50) NULL,
    payment_gateway VARCHAR(50) NULL,
    transaction_id VARCHAR(100) NULL,
    payment_status VARCHAR(20) NULL,
    paid_at TIMESTAMP NULL,
    
    -- Billing period
    billing_period_start TIMESTAMP NULL,
    billing_period_end TIMESTAMP NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    UNIQUE KEY uk_order_number (order_number),
    INDEX idx_orders_user (user_id, status),
    INDEX idx_orders_plan (subscription_plan_id),
    INDEX idx_orders_payment (payment_status, paid_at),
    INDEX idx_orders_created (created_at),
    INDEX idx_orders_deleted (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- SUPPORT TICKETS
-- ==============================================================================

CREATE TABLE tickets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ticket_no VARCHAR(32) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Ticket details
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    
    -- Assignment
    assigned_to_id BIGINT UNSIGNED NULL,
    assigned_at TIMESTAMP NULL,
    
    -- Resolution
    resolved_by_id BIGINT UNSIGNED NULL,
    resolved_at TIMESTAMP NULL,
    resolution TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    UNIQUE KEY uk_ticket_no (ticket_no),
    INDEX idx_tickets_user (user_id),
    INDEX idx_tickets_status (status),
    INDEX idx_tickets_assigned (assigned_to_id),
    INDEX idx_tickets_created (created_at),
    INDEX idx_tickets_deleted (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE ticket_messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ticket_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Message details
    content TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'user',
    is_internal BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_ticket_msgs_ticket (ticket_id, created_at),
    INDEX idx_ticket_msgs_user (user_id),
    INDEX idx_ticket_msgs_deleted (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- REFERRAL SYSTEM
-- ==============================================================================

CREATE TABLE referral_campaigns (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT NULL,
    
    -- Campaign settings
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    start_date TIMESTAMP NULL,
    end_date TIMESTAMP NULL,
    
    -- Rewards
    referrer_reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    referee_reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    reward_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Stats
    total_referrals INT NOT NULL DEFAULT 0,
    total_rewards_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    UNIQUE KEY uk_campaign_code (code),
    INDEX idx_campaigns_status (status),
    INDEX idx_campaigns_created (created_at),
    INDEX idx_campaigns_deleted (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE referrals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    referrer_id BIGINT UNSIGNED NOT NULL,
    referee_id BIGINT UNSIGNED NOT NULL,
    campaign_id BIGINT UNSIGNED NULL,
    
    -- Tracking
    referral_code VARCHAR(100) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reward_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- Rewards
    reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    referee_reward DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    reward_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Timestamps
    converted_at TIMESTAMP NULL,
    rewarded_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_referrals_referrer (referrer_id),
    INDEX idx_referrals_referee (referee_id),
    INDEX idx_referrals_campaign (campaign_id),
    INDEX idx_referrals_status (status),
    INDEX idx_referrals_created (created_at),
    INDEX idx_referrals_deleted (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- SERVER GROUPS - SERVER ORGANIZATION
-- ==============================================================================

CREATE TABLE server_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT NULL,
    sort INT DEFAULT 0,
    is_show BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    INDEX idx_server_groups_name (name),
    INDEX idx_server_groups_sort (sort),
    INDEX idx_server_groups_created_at (created_at),
    INDEX idx_server_groups_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- SHADOWSOCKS SERVERS
-- ==============================================================================

CREATE TABLE shadowsocks_servers (
    id INT NOT NULL AUTO_INCREMENT,
    group_id VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    route_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    parent_id INT DEFAULT NULL,
    tags VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    excludes TEXT COLLATE utf8mb4_unicode_ci,
    ips VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    name VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    rate VARCHAR(11) COLLATE utf8mb4_unicode_ci NOT NULL,
    host VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    port VARCHAR(11) COLLATE utf8mb4_unicode_ci NOT NULL,
    server_port INT NOT NULL,
    cipher VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    obfs CHAR(11) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    obfs_settings VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `show` TINYINT NOT NULL DEFAULT '0',
    sort INT DEFAULT NULL,
    created_at INT NOT NULL,
    updated_at INT NOT NULL,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- USER SUBSCRIPTION SERVER GROUPS - MANY-TO-MANY RELATIONSHIP
-- ==============================================================================

CREATE TABLE user_subscription_server_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    server_group_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_subscription_id (user_subscription_id),
    INDEX idx_server_group_id (server_group_id),
    UNIQUE KEY unique_subscription_group (user_subscription_id, server_group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- PAYMENT CONFIGURATIONS
-- ==============================================================================

CREATE TABLE payment_configs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    gateway VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    
    -- Configuration settings (JSON)
    config JSON NOT NULL,
    
    -- Supported currencies and methods
    supported_currencies TEXT NULL,
    supported_methods TEXT NULL,
    
    -- Limits
    min_amount DECIMAL(10,2) NOT NULL DEFAULT 0.01,
    max_amount DECIMAL(10,2) NOT NULL DEFAULT 99999.99,
    
    -- Fees
    fixed_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    percentage_fee DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    UNIQUE KEY uk_payment_config_gateway (gateway),
    INDEX idx_payment_configs_enabled (is_enabled, sort_order),
    INDEX idx_payment_configs_deleted (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- INVOICES TABLE
-- ==============================================================================

CREATE TABLE invoices (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_order_id BIGINT UNSIGNED NOT NULL,
    
    -- Invoice Information
    invoice_number VARCHAR(50) NOT NULL,
    invoice_type VARCHAR(20) NOT NULL DEFAULT 'standard',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    
    -- Financial Details
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    tax_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    
    -- Tax Information
    tax_rate DECIMAL(5,4) NOT NULL DEFAULT 0,
    tax_type VARCHAR(20) NULL,
    tax_number VARCHAR(50) NULL,
    
    -- Billing Information
    billing_name VARCHAR(200) NOT NULL,
    billing_email VARCHAR(191) NOT NULL,
    billing_address TEXT NULL,
    billing_city VARCHAR(100) NULL,
    billing_state VARCHAR(100) NULL,
    billing_country VARCHAR(2) NULL,
    billing_zip VARCHAR(20) NULL,
    
    -- Company Information
    company_name VARCHAR(200) NULL,
    company_tax_id VARCHAR(50) NULL,
    company_address TEXT NULL,
    
    -- Important Dates
    issued_at TIMESTAMP NOT NULL,
    due_at TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,
    voided_at TIMESTAMP NULL,
    
    -- Payment Information
    payment_method VARCHAR(50) NULL,
    payment_reference VARCHAR(100) NULL,
    
    -- Invoice Template and Language
    template VARCHAR(50) NULL DEFAULT 'default',
    language VARCHAR(5) NULL DEFAULT 'en',
    
    -- File Storage
    pdf_path VARCHAR(500) NULL,
    pdf_size BIGINT NULL,
    
    -- Additional Information
    description TEXT NULL,
    notes TEXT NULL,
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    UNIQUE KEY uk_invoice_number (invoice_number),
    INDEX idx_invoices_user (user_id),
    INDEX idx_invoices_order (subscription_order_id),
    INDEX idx_invoices_status (status),
    INDEX idx_invoices_type (invoice_type),
    INDEX idx_invoices_issued (issued_at),
    INDEX idx_invoices_due (due_at),
    INDEX idx_invoices_paid (paid_at),
    INDEX idx_invoices_sent (sent_at),
    INDEX idx_invoices_created (created_at),
    INDEX idx_invoices_deleted (deleted_at),
    
    -- Composite indexes for common queries
    INDEX idx_invoices_user_status (user_id, status),
    INDEX idx_invoices_status_due (status, due_at),
    INDEX idx_invoices_currency_amount (currency, total_amount)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ==============================================================================
-- SECURITY FEATURES
-- ==============================================================================

-- JWT Blacklist table for token revocation
CREATE TABLE jwt_blacklist (
    token_hash VARCHAR(64) PRIMARY KEY COMMENT 'SHA256 hash of the JWT token',
    user_id INT UNSIGNED NULL COMMENT 'User ID (application-level constraint)',
    reason VARCHAR(100) NOT NULL COMMENT 'Reason for blacklisting (logout, security_breach, etc.)',
    expires_at TIMESTAMP NOT NULL COMMENT 'When the token expires (no need to check after this)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    INDEX idx_jwt_blacklist_user_id (user_id),
    INDEX idx_jwt_blacklist_expires_at (expires_at),
    INDEX idx_jwt_blacklist_created_at (created_at),
    INDEX idx_jwt_blacklist_reason (reason)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='JWT token blacklist for secure logout and token revocation';

-- Login attempts table for security tracking
CREATE TABLE login_attempts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    ip VARCHAR(45) NOT NULL COMMENT 'IPv4 or IPv6 address',
    user_agent VARCHAR(500) NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    reason VARCHAR(200) NULL COMMENT 'Success/failure reason',
    user_id INT UNSIGNED NULL COMMENT 'User ID (application-level constraint)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    INDEX idx_login_attempts_email (email),
    INDEX idx_login_attempts_ip (ip),
    INDEX idx_login_attempts_success (success),
    INDEX idx_login_attempts_created_at (created_at),
    INDEX idx_login_attempts_user_id (user_id),
    INDEX idx_login_attempts_email_success_created (email, success, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Login attempt tracking for security analysis and lockout management';

-- Account lockout table for managing locked accounts
CREATE TABLE account_lockouts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    user_id INT UNSIGNED NULL COMMENT 'User ID (application-level constraint)',
    failed_count INT NOT NULL DEFAULT 0 COMMENT 'Number of failed attempts in current window',
    last_failure TIMESTAMP NULL COMMENT 'Time of last failed attempt',
    locked_until TIMESTAMP NULL COMMENT 'Account locked until this time (NULL = not locked)',
    lock_reason VARCHAR(200) NULL COMMENT 'Reason for lock',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    
    INDEX idx_account_lockouts_email (email),
    INDEX idx_account_lockouts_user_id (user_id),
    INDEX idx_account_lockouts_locked_until (locked_until),
    INDEX idx_account_lockouts_last_failure (last_failure)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Account lockout management for failed login attempts';

-- Add indexes for performance optimization
-- Composite index for efficient blacklist cleanup
CREATE INDEX idx_jwt_blacklist_cleanup ON jwt_blacklist (expires_at, created_at);

-- Composite index for login attempt analysis
CREATE INDEX idx_login_attempts_analysis ON login_attempts (email, success, created_at);

-- Index for IP-based analysis
CREATE INDEX idx_login_attempts_ip_analysis ON login_attempts (ip, success, created_at);

-- ==============================================================================
-- DATA INITIALIZATION
-- ==============================================================================

-- Initialize subscription plans with proper data
INSERT INTO subscription_plans (id, name, code, description, price, currency, billing_cycle, billing_interval, status, is_visible, created_at, updated_at) VALUES
(1, 'Basic Plan', 'basic', 'Basic subscription plan', 9.99, 'USD', 'monthly', 1, 'active', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Set proper default values for any existing user subscriptions
UPDATE user_subscriptions 
SET uuid = UUID() 
WHERE uuid IS NULL OR uuid = '' OR LENGTH(uuid) != 36;

-- Ensure subscription plans have proper status
UPDATE subscription_plans 
SET 
    status = 'active',
    is_visible = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE status = '' OR status IS NULL;