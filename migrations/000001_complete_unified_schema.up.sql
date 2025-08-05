-- ==============================================================================
-- COMPREHENSIVE UNIFIED DATABASE SCHEMA - COMPLETE MIGRATION
-- ==============================================================================
-- This migration creates the complete unified database schema for the Linke platform
-- Consolidates all existing migrations into a single optimized migration following:
-- - MySQL best practices with optimized indexing strategies
-- - Native MySQL data types and constraints  
-- - Performance-optimized table structures
-- - No foreign key constraints (application-level enforcement)
-- - Event-driven architecture support with comprehensive event store
-- - Multi-level caching optimization support
-- - Complete referral system with rewards and tracking
-- - Advanced payment processing with retry mechanisms
-- - Comprehensive usage tracking and alerting system
-- ==============================================================================

-- ==============================================================================
-- CORE USER MANAGEMENT
-- ==============================================================================

-- Users table with complete authentication and profile management
CREATE TABLE users (
    -- Primary key
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Core authentication fields
    email VARCHAR(191) NOT NULL,
    password VARCHAR(255) NULL,
    
    -- Status and role management
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    provider VARCHAR(20) NOT NULL DEFAULT 'local',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Profile information
    username VARCHAR(32) NULL,
    name VARCHAR(128) NULL,
    phone VARCHAR(20) NULL,
    avatar VARCHAR(255) NULL,
    
    -- OAuth provider integration
    google_id VARCHAR(32) NULL,
    github_id VARCHAR(32) NULL,
    telegram_id VARCHAR(32) NULL,
    provider_data JSON NULL,
    
    -- Security and audit fields
    failed_login_attempts TINYINT UNSIGNED NOT NULL DEFAULT 0,
    login_count INT UNSIGNED NOT NULL DEFAULT 0,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45) NULL,
    locked_until TIMESTAMP NULL,
    
    -- Verification timestamps
    email_verified_at TIMESTAMP NULL,
    phone_verified_at TIMESTAMP NULL,
    
    -- Referral system integration
    total_referrals SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    total_referral_rewards DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    invite_code_id BIGINT UNSIGNED NULL,
    invite_code_used VARCHAR(32) NULL,
    referral_code VARCHAR(16) NULL,
    referral_link VARCHAR(255) NULL,
    
    -- Standard timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential unique constraints
    UNIQUE KEY uk_users_email (email),
    UNIQUE KEY uk_users_username (username),
    UNIQUE KEY uk_users_google_id (google_id),
    UNIQUE KEY uk_users_github_id (github_id),
    UNIQUE KEY uk_users_telegram_id (telegram_id),
    
    -- Performance indexes
    INDEX idx_users_status_role (status, role),
    INDEX idx_users_provider (provider),
    INDEX idx_users_referral_code (referral_code),
    INDEX idx_users_created_at (created_at),
    INDEX idx_users_deleted_at (deleted_at),
    INDEX idx_users_last_login (last_login_at),
    
    -- Covering index for common queries
    INDEX idx_users_covering (id, email, username, status, role, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Core user management with OAuth support and referral system';

-- ==============================================================================
-- SECURITY AND AUTHENTICATION
-- ==============================================================================

-- JWT blacklist for secure token revocation
CREATE TABLE jwt_blacklist (
    token_hash VARCHAR(64) NOT NULL PRIMARY KEY COMMENT 'SHA256 hash of JWT token',
    user_id BIGINT UNSIGNED NULL,
    reason VARCHAR(100) NOT NULL COMMENT 'Blacklist reason',
    expires_at TIMESTAMP NOT NULL COMMENT 'Token expiration time',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_jwt_blacklist_user_id (user_id),
    INDEX idx_jwt_blacklist_expires_at (expires_at),
    INDEX idx_jwt_blacklist_reason (reason),
    INDEX idx_jwt_blacklist_cleanup (expires_at, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='JWT token blacklist for secure logout and revocation';

-- Login attempt tracking for security
CREATE TABLE login_attempts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email VARCHAR(191) NOT NULL,
    ip VARCHAR(45) NOT NULL,
    user_agent VARCHAR(500) NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    reason VARCHAR(200) NULL,
    user_id BIGINT UNSIGNED NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    INDEX idx_login_attempts_email (email),
    INDEX idx_login_attempts_ip (ip),
    INDEX idx_login_attempts_success (success),
    INDEX idx_login_attempts_created_at (created_at),
    INDEX idx_login_attempts_user_id (user_id),
    INDEX idx_login_attempts_analysis (email, success, created_at),
    INDEX idx_login_attempts_ip_analysis (ip, success, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Login attempt tracking for security analysis';

-- Account lockout management
CREATE TABLE account_lockouts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email VARCHAR(191) NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    failed_count INT NOT NULL DEFAULT 0,
    last_failure TIMESTAMP NULL,
    locked_until TIMESTAMP NULL,
    lock_reason VARCHAR(200) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_account_lockouts_email (email),
    INDEX idx_account_lockouts_user_id (user_id),
    INDEX idx_account_lockouts_locked_until (locked_until),
    INDEX idx_account_lockouts_last_failure (last_failure)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Account lockout management for failed login attempts';

-- ==============================================================================
-- SUBSCRIPTION MANAGEMENT
-- ==============================================================================

-- Subscription plans with traffic management
CREATE TABLE subscription_plans (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT NULL,
    
    -- Pricing configuration
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_cycle VARCHAR(20) NOT NULL,
    billing_interval INT NOT NULL DEFAULT 1,
    
    -- Plan features and limits
    features TEXT NULL,
    limits TEXT NULL,
    
    -- Display and sorting
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    is_popular BOOLEAN NOT NULL DEFAULT FALSE,
    is_recommended BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    
    -- Additional fees
    setup_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    cancellation_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    trial_period_days INT NOT NULL DEFAULT 0,
    
    -- Traffic configuration
    traffic_limit BIGINT NOT NULL DEFAULT 10737418240 COMMENT 'Traffic limit in bytes (default: 10GB)',
    traffic_reset_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    
    -- Server group defaults (JSON array)
    default_server_groups TEXT NULL COMMENT 'JSON array of default server group IDs',
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_subscription_plans_code (code),
    INDEX idx_subscription_plans_status (status, is_visible),
    INDEX idx_subscription_plans_currency (currency),
    INDEX idx_subscription_plans_sort (sort_order),
    INDEX idx_subscription_plans_popular (is_popular, is_recommended),
    INDEX idx_subscription_plans_deleted_at (deleted_at),
    
    -- Covering index for plan listing
    INDEX idx_subscription_plans_covering (id, name, price, status, is_visible, sort_order)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Subscription plans with traffic and server group configuration';

-- User subscriptions with comprehensive management
CREATE TABLE user_subscriptions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    uuid VARCHAR(36) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_plan_id BIGINT UNSIGNED NOT NULL,
    
    -- Subscription status and configuration
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    is_trial BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Pricing (locked at creation)
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_cycle VARCHAR(20) NOT NULL,
    billing_interval INT NOT NULL DEFAULT 1,
    
    -- Service period management
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NULL,
    current_period_start TIMESTAMP NULL,
    current_period_end TIMESTAMP NULL,
    next_billing_date TIMESTAMP NULL,
    trial_end_date TIMESTAMP NULL,
    
    -- Cancellation management
    cancelled_at TIMESTAMP NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    cancellation_reason VARCHAR(255) NULL,
    
    -- Pause and resume functionality
    is_paused BOOLEAN NOT NULL DEFAULT FALSE,
    paused_at TIMESTAMP NULL,
    pause_reason VARCHAR(255) NULL,
    resume_at TIMESTAMP NULL,
    paused_cycles INT NOT NULL DEFAULT 0,
    max_pause_cycles INT NOT NULL DEFAULT 3,
    
    -- Payment tracking
    payment_failed_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
    last_payment_attempt TIMESTAMP NULL,
    last_used_at TIMESTAMP NULL,
    
    -- Traffic management
    traffic_limit BIGINT NOT NULL DEFAULT 0 COMMENT 'Traffic limit in bytes',
    traffic_used BIGINT NOT NULL DEFAULT 0 COMMENT 'Traffic used in bytes',
    traffic_reset_date TIMESTAMP NULL COMMENT 'Next traffic reset date',
    traffic_reset_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    traffic_suspended BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Server group access (JSON array)
    server_group_ids TEXT NULL COMMENT 'JSON array of accessible server group IDs',
    
    -- Additional data
    usage_data TEXT NULL,
    notes TEXT NULL,
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential unique constraints
    UNIQUE KEY uk_user_subscriptions_uuid (uuid),
    
    -- Performance indexes
    INDEX idx_user_subscriptions_user (user_id, status),
    INDEX idx_user_subscriptions_plan (subscription_plan_id),
    INDEX idx_user_subscriptions_billing (next_billing_date, is_auto_renew),
    INDEX idx_user_subscriptions_active (is_active, status),
    INDEX idx_user_subscriptions_trial (is_trial, trial_end_date),
    INDEX idx_user_subscriptions_pause (is_paused, resume_at),
    INDEX idx_user_subscriptions_traffic (traffic_suspended, status),
    INDEX idx_user_subscriptions_deleted_at (deleted_at),
    
    -- Composite indexes for common queries
    INDEX idx_user_subscriptions_user_status_created (user_id, status, created_at DESC),
    INDEX idx_user_subscriptions_plan_status_created (subscription_plan_id, status, created_at DESC),
    INDEX idx_user_subscriptions_traffic_management (traffic_limit, traffic_used, traffic_reset_date)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='User subscriptions with comprehensive lifecycle management';

-- ==============================================================================
-- ORDER AND BILLING SYSTEM
-- ==============================================================================

-- Subscription orders for purchase workflow
CREATE TABLE subscription_orders (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_number VARCHAR(50) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_plan_id BIGINT UNSIGNED NOT NULL,
    user_subscription_id BIGINT UNSIGNED NULL,
    
    -- Order configuration
    order_type VARCHAR(20) NOT NULL DEFAULT 'new',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- Financial details
    amount DECIMAL(10,2) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    tax_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Refund management
    refund_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    refund_status VARCHAR(20) NULL,
    refunded_at TIMESTAMP NULL,
    refund_reason VARCHAR(100) NULL,
    
    -- Coupon integration
    coupon_code VARCHAR(50) NULL,
    coupon_discount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Payment information
    payment_method VARCHAR(50) NULL,
    payment_gateway VARCHAR(50) NULL,
    transaction_id VARCHAR(100) NULL,
    payment_status VARCHAR(20) NULL,
    paid_at TIMESTAMP NULL,
    
    -- Service period
    billing_period_start TIMESTAMP NULL,
    billing_period_end TIMESTAMP NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_subscription_orders_order_number (order_number),
    INDEX idx_subscription_orders_user (user_id, status),
    INDEX idx_subscription_orders_plan (subscription_plan_id),
    INDEX idx_subscription_orders_subscription (user_subscription_id),
    INDEX idx_subscription_orders_payment (payment_status, paid_at),
    INDEX idx_subscription_orders_refund (refund_status, refunded_at),
    INDEX idx_subscription_orders_coupon (coupon_code),
    INDEX idx_subscription_orders_created_at (created_at),
    INDEX idx_subscription_orders_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Subscription purchase orders with comprehensive billing';

-- ==============================================================================
-- PAYMENT PROCESSING
-- ==============================================================================

-- Payment records with comprehensive tracking
CREATE TABLE payment_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    payment_no CHAR(32) NOT NULL,
    out_trade_no CHAR(32) NOT NULL,
    transaction_id VARCHAR(64) NULL,
    
    -- References
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_order_id BIGINT UNSIGNED NULL,
    invoice_id BIGINT UNSIGNED NULL,
    
    -- Payment details
    gateway VARCHAR(30) NOT NULL,
    payment_method VARCHAR(30) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(20) NULL,
    
    -- Security fields for anti-replay protection
    idempotency_key VARCHAR(255) NULL COMMENT 'Anti-replay protection key',
    client_fingerprint VARCHAR(255) NULL COMMENT 'Client device fingerprint',
    security_hash VARCHAR(255) NULL COMMENT 'Security verification hash',
    
    -- Important timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    paid_at TIMESTAMP NULL,
    expired_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    
    -- Refund management
    refund_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    refund_status VARCHAR(20) NULL,
    refunded_at TIMESTAMP NULL,
    refund_reason VARCHAR(100) NULL,
    
    -- Retry tracking
    retry_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
    last_retry_at TIMESTAMP NULL,
    
    -- Client context
    client_ip VARCHAR(45) NULL,
    user_agent TEXT NULL,
    
    -- URLs and gateway response
    payment_url VARCHAR(512) NULL,
    notify_url VARCHAR(512) NULL,
    return_url VARCHAR(512) NULL,
    gateway_response JSON NULL,
    
    PRIMARY KEY (id),
    
    -- Essential unique constraints
    UNIQUE KEY uk_payment_records_payment_no (payment_no),
    UNIQUE KEY uk_payment_records_out_trade_no (out_trade_no),
    UNIQUE KEY uk_payment_records_idempotency (idempotency_key),
    
    -- Performance indexes
    INDEX idx_payment_records_user (user_id, status),
    INDEX idx_payment_records_order (subscription_order_id),
    INDEX idx_payment_records_invoice (invoice_id),
    INDEX idx_payment_records_gateway (gateway, status),
    INDEX idx_payment_records_status (status, paid_at),
    INDEX idx_payment_records_transaction (transaction_id),
    INDEX idx_payment_records_created_at (created_at),
    INDEX idx_payment_records_deleted_at (deleted_at),
    
    -- Security and anti-fraud indexes
    INDEX idx_payment_records_security (client_fingerprint, client_ip),
    INDEX idx_payment_records_retry (retry_count, last_retry_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Payment records with security and anti-fraud protection';

-- Payment retry system for smart retry strategy
CREATE TABLE payment_retries (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    
    -- Retry configuration
    attempt_number INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP NULL,
    last_attempt_at TIMESTAMP NULL,
    retry_strategy VARCHAR(50) NOT NULL DEFAULT 'exponential',
    
    -- Status and failure tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    failure_type VARCHAR(30) NULL COMMENT 'temporary, permanent, network, gateway, business',
    last_failure_code VARCHAR(50) NULL,
    last_error_message VARCHAR(500) NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    INDEX idx_payment_retries_payment_record (payment_record_id),
    INDEX idx_payment_retries_status (status),
    INDEX idx_payment_retries_next_retry (next_retry_at),
    INDEX idx_payment_retries_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Smart payment retry tracking';

-- Payment retry history for audit trail
CREATE TABLE payment_retry_histories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    payment_retry_id BIGINT UNSIGNED NOT NULL,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    
    -- Attempt details
    attempt_number INT NOT NULL,
    attempted_at TIMESTAMP NULL,
    status VARCHAR(20) NOT NULL,
    error_message VARCHAR(500) NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    INDEX idx_payment_retry_histories_retry (payment_retry_id),
    INDEX idx_payment_retry_histories_payment (payment_record_id),
    INDEX idx_payment_retry_histories_attempted (attempted_at),
    INDEX idx_payment_retry_histories_status (status),
    INDEX idx_payment_retry_histories_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Payment retry attempt audit trail';

-- Payment methods for secure tokenized storage
CREATE TABLE payment_methods (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Method classification
    type VARCHAR(50) NOT NULL COMMENT 'card, bank_account, digital_wallet, crypto',
    gateway VARCHAR(50) NOT NULL,
    method VARCHAR(50) NOT NULL COMMENT 'alipay, wechat, usdt, etc.',
    display_name VARCHAR(100) NOT NULL,
    
    -- Tokenized payment data (PCI compliant)
    payment_token VARCHAR(255) NOT NULL,
    gateway_customer_id VARCHAR(255) NULL,
    
    -- Safe display information
    masked_info VARCHAR(100) NULL COMMENT 'e.g., "**** 1234", "ali***@example.com"',
    brand VARCHAR(50) NULL,
    expiry_month INT NULL,
    expiry_year INT NULL,
    
    -- Status and configuration
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Security and validation
    last_validated_at TIMESTAMP NULL,
    validation_hash VARCHAR(64) NULL,
    
    -- Billing information
    billing_country VARCHAR(10) NULL,
    billing_postcode VARCHAR(20) NULL,
    
    -- Gateway metadata
    gateway_metadata TEXT NULL,
    
    -- Usage statistics
    last_used_at TIMESTAMP NULL,
    successful_uses INT NOT NULL DEFAULT 0,
    failed_uses INT NOT NULL DEFAULT 0,
    
    -- Security tracking
    created_from_ip VARCHAR(45) NULL,
    last_update_ip VARCHAR(45) NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_payment_methods_user (user_id),
    INDEX idx_payment_methods_gateway (gateway),
    INDEX idx_payment_methods_type (type),
    INDEX idx_payment_methods_status (status),
    INDEX idx_payment_methods_default (is_default),
    INDEX idx_payment_methods_active (is_active),
    INDEX idx_payment_methods_token (payment_token),
    INDEX idx_payment_methods_deleted_at (deleted_at),
    
    -- Unique constraints
    UNIQUE KEY uk_payment_methods_gateway_token (gateway, payment_token)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Tokenized payment methods for PCI compliance';

-- Payment configurations for gateway management
CREATE TABLE payment_configs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    gateway VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    
    -- Configuration (JSON)
    config JSON NOT NULL,
    
    -- Supported options
    supported_currencies TEXT NULL,
    supported_methods TEXT NULL,
    
    -- Limits and fees
    min_amount DECIMAL(10,2) NOT NULL DEFAULT 0.01,
    max_amount DECIMAL(10,2) NOT NULL DEFAULT 99999.99,
    fixed_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    percentage_fee DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_payment_configs_gateway (gateway),
    INDEX idx_payment_configs_enabled (is_enabled, sort_order),
    INDEX idx_payment_configs_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Payment gateway configurations';

-- ==============================================================================
-- INVOICE MANAGEMENT
-- ==============================================================================

-- Comprehensive invoice system
CREATE TABLE invoices (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_order_id BIGINT UNSIGNED NOT NULL,
    
    -- Invoice identification
    invoice_number VARCHAR(50) NOT NULL,
    invoice_type VARCHAR(20) NOT NULL DEFAULT 'standard',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    
    -- Financial details
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    tax_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(10,2) NOT NULL,
    
    -- Tax information
    tax_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    tax_type VARCHAR(20) NULL,
    tax_number VARCHAR(50) NULL,
    
    -- Billing information
    billing_name VARCHAR(200) NOT NULL,
    billing_email VARCHAR(191) NOT NULL,
    billing_address TEXT NULL,
    billing_city VARCHAR(100) NULL,
    billing_state VARCHAR(100) NULL,
    billing_country VARCHAR(2) NULL,
    billing_zip VARCHAR(20) NULL,
    
    -- Company information
    company_name VARCHAR(200) NULL,
    company_tax_id VARCHAR(50) NULL,
    company_address TEXT NULL,
    
    -- Important dates
    issued_at TIMESTAMP NOT NULL,
    due_at TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,
    voided_at TIMESTAMP NULL,
    
    -- Payment information
    payment_method VARCHAR(50) NULL,
    payment_reference VARCHAR(100) NULL,
    
    -- Template and language
    template VARCHAR(50) NULL DEFAULT 'default',
    language VARCHAR(5) NULL DEFAULT 'en',
    
    -- File management
    pdf_path VARCHAR(500) NULL,
    pdf_size BIGINT NULL,
    
    -- Content
    description TEXT NULL,
    notes TEXT NULL,
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_invoices_invoice_number (invoice_number),
    INDEX idx_invoices_user (user_id),
    INDEX idx_invoices_order (subscription_order_id),
    INDEX idx_invoices_status (status),
    INDEX idx_invoices_type (invoice_type),
    INDEX idx_invoices_issued (issued_at),
    INDEX idx_invoices_due (due_at),
    INDEX idx_invoices_paid (paid_at),
    INDEX idx_invoices_sent (sent_at),
    INDEX idx_invoices_deleted_at (deleted_at),
    
    -- Composite indexes for common queries
    INDEX idx_invoices_user_status (user_id, status),
    INDEX idx_invoices_status_due (status, due_at),
    INDEX idx_invoices_currency_amount (currency, total_amount)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Comprehensive invoice management system';

-- ==============================================================================
-- USAGE TRACKING AND ALERTS
-- ==============================================================================

-- High-performance usage tracking
CREATE TABLE usage_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    
    -- Usage details
    usage_type VARCHAR(50) NOT NULL COMMENT 'traffic, api_calls, storage, etc.',
    amount BIGINT NOT NULL COMMENT 'Usage amount in bytes or count',
    unit VARCHAR(20) NOT NULL DEFAULT 'bytes',
    timestamp TIMESTAMP NOT NULL,
    
    -- Source information
    source_type VARCHAR(50) NOT NULL COMMENT 'server, api, admin, etc.',
    source_id VARCHAR(100) NULL,
    
    -- Additional context
    metadata TEXT NULL,
    user_agent TEXT NULL,
    ip_address VARCHAR(45) NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- High-performance time-series indexes
    INDEX idx_usage_records_subscription_type_time (user_subscription_id, usage_type, timestamp DESC),
    INDEX idx_usage_records_timestamp (timestamp),
    INDEX idx_usage_records_usage_type (usage_type, timestamp),
    INDEX idx_usage_records_source (source_type, source_id),
    INDEX idx_usage_records_deleted_at (deleted_at),
    
    -- Covering index for reporting
    INDEX idx_usage_records_covering (user_subscription_id, usage_type, timestamp, amount, unit)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='High-performance usage tracking for real-time monitoring';

-- Alert configurations for threshold monitoring
CREATE TABLE alert_configurations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    
    -- Alert settings
    usage_type VARCHAR(50) NOT NULL,
    threshold_type VARCHAR(20) NOT NULL DEFAULT 'percentage',
    threshold DECIMAL(10,4) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Notification settings
    notification_channels TEXT NULL COMMENT 'JSON array of notification channels',
    cooldown_minutes INT NOT NULL DEFAULT 60,
    
    -- Metadata
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_alert_configurations_subscription (user_subscription_id, is_enabled),
    INDEX idx_alert_configurations_usage_type (usage_type, is_enabled),
    INDEX idx_alert_configurations_enabled (is_enabled, deleted_at),
    INDEX idx_alert_configurations_priority (priority),
    INDEX idx_alert_configurations_deleted_at (deleted_at),
    
    -- Threshold lookup optimization
    INDEX idx_alert_configurations_lookup (user_subscription_id, usage_type, is_enabled, 
        threshold_type, threshold)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='User-configured alert thresholds for usage monitoring';

-- Usage alerts for notification tracking
CREATE TABLE usage_alerts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    alert_configuration_id BIGINT UNSIGNED NOT NULL,
    
    -- Alert details
    usage_type VARCHAR(50) NOT NULL,
    current_usage BIGINT NOT NULL,
    usage_limit BIGINT NOT NULL,
    threshold_value DECIMAL(10,4) NOT NULL,
    usage_percent DECIMAL(5,2) NOT NULL,
    
    -- Alert state
    status VARCHAR(20) NOT NULL DEFAULT 'fired',
    severity VARCHAR(20) NOT NULL,
    fired_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP NULL,
    
    -- Notification tracking
    notifications_sent INT NOT NULL DEFAULT 0,
    last_notification_sent TIMESTAMP NULL,
    notification_channels TEXT NULL,
    notification_results TEXT NULL,
    
    -- Additional context
    message TEXT NULL,
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_usage_alerts_subscription (user_subscription_id, status),
    INDEX idx_usage_alerts_configuration (alert_configuration_id),
    INDEX idx_usage_alerts_usage_type (usage_type, status),
    INDEX idx_usage_alerts_status (status, resolved_at),
    INDEX idx_usage_alerts_fired_at (fired_at),
    INDEX idx_usage_alerts_severity (severity, status),
    INDEX idx_usage_alerts_deleted_at (deleted_at),
    
    -- Dashboard optimization indexes
    INDEX idx_usage_alerts_active (user_subscription_id, status, fired_at DESC),
    INDEX idx_usage_alerts_recent (status, fired_at DESC, resolved_at),
    INDEX idx_usage_alerts_notifications (last_notification_sent, status),
    
    -- Covering index for alert lists
    INDEX idx_usage_alerts_covering (user_subscription_id, status, fired_at, severity, usage_type)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Fired usage alerts and notification tracking';

-- ==============================================================================
-- COUPON SYSTEM
-- ==============================================================================

-- Comprehensive coupon management
CREATE TABLE coupons (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Core coupon information
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    type VARCHAR(20) NOT NULL COMMENT 'percentage, fixed_amount',
    value DECIMAL(10,2) NOT NULL,
    
    -- Usage limits
    max_uses INT NOT NULL DEFAULT 1,
    used_count INT NOT NULL DEFAULT 0,
    max_uses_per_user INT NOT NULL DEFAULT 1,
    
    -- Order requirements
    min_order_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Validity period
    valid_from TIMESTAMP NULL,
    valid_until TIMESTAMP NULL,
    
    -- Applicable plans (JSON array)
    applicable_plans TEXT NULL COMMENT 'JSON array of applicable plan IDs',
    
    -- Status and visibility
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT UNSIGNED NOT NULL,
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_coupons_code (code),
    INDEX idx_coupons_status (status),
    INDEX idx_coupons_type (type),
    INDEX idx_coupons_public (is_public),
    INDEX idx_coupons_created_by (created_by),
    INDEX idx_coupons_valid_from (valid_from),
    INDEX idx_coupons_valid_until (valid_until),
    INDEX idx_coupons_deleted_at (deleted_at),
    
    -- Composite indexes for performance
    INDEX idx_coupons_status_public (status, is_public),
    INDEX idx_coupons_valid_period (valid_from, valid_until)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Comprehensive coupon management system';

-- Coupon usage tracking and fraud prevention
CREATE TABLE coupon_usages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- References
    coupon_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_order_id BIGINT UNSIGNED NOT NULL,
    
    -- Usage details
    discount_amount DECIMAL(10,2) NOT NULL,
    order_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_coupon_usages_coupon (coupon_id),
    INDEX idx_coupon_usages_user (user_id),
    INDEX idx_coupon_usages_order (subscription_order_id),
    INDEX idx_coupon_usages_created_at (created_at),
    INDEX idx_coupon_usages_deleted_at (deleted_at),
    
    -- Composite indexes
    INDEX idx_coupon_usages_coupon_user (coupon_id, user_id),
    INDEX idx_coupon_usages_user_created (user_id, created_at),
    INDEX idx_coupon_usages_stats (coupon_id, created_at, discount_amount),
    
    -- Unique constraint to prevent duplicate usage
    UNIQUE KEY uk_coupon_usages_order (coupon_id, subscription_order_id)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Coupon usage tracking and fraud prevention';

-- ==============================================================================
-- COMPLETE REFERRAL SYSTEM
-- ==============================================================================

-- Invite codes for referral system
CREATE TABLE invite_codes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code CHAR(32) NOT NULL,
    created_by_id BIGINT UNSIGNED NOT NULL,
    
    -- Status and configuration
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
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_invite_codes_code (code),
    INDEX idx_invite_codes_creator (created_by_id),
    INDEX idx_invite_codes_active (is_active, status, expires_at),
    INDEX idx_invite_codes_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Invite codes for referral system';

-- Invite code usage audit trail
CREATE TABLE invite_code_usages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    invite_code_id BIGINT UNSIGNED NOT NULL,
    used_by_id BIGINT UNSIGNED NOT NULL,
    used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(255) NULL,
    
    PRIMARY KEY (id),
    INDEX idx_invite_code_usages_code (invite_code_id),
    INDEX idx_invite_code_usages_user (used_by_id),
    INDEX idx_invite_code_usages_time (used_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Invite code usage audit trail';

-- Referral campaigns
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
    
    -- Statistics
    total_referrals INT NOT NULL DEFAULT 0,
    total_rewards_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_referral_campaigns_code (code),
    INDEX idx_referral_campaigns_status (status),
    INDEX idx_referral_campaigns_created_at (created_at),
    INDEX idx_referral_campaigns_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Referral campaigns for reward management';

-- Referral tracking
CREATE TABLE referrals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    referrer_id BIGINT UNSIGNED NOT NULL,
    referee_id BIGINT UNSIGNED NOT NULL,
    campaign_id BIGINT UNSIGNED NULL,
    
    -- Tracking information
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
    INDEX idx_referrals_created_at (created_at),
    INDEX idx_referrals_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Referral tracking and reward management';

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

-- ==============================================================================
-- SERVER MANAGEMENT
-- ==============================================================================

-- Server groups for organization
CREATE TABLE server_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    sort INT NOT NULL DEFAULT 0,
    is_show BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    INDEX idx_server_groups_name (name),
    INDEX idx_server_groups_sort (sort),
    INDEX idx_server_groups_show (is_show),
    INDEX idx_server_groups_created_at (created_at),
    INDEX idx_server_groups_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Server groups for organization and access control';

-- Shadowsocks servers configuration
CREATE TABLE shadowsocks_servers (
    id INT NOT NULL AUTO_INCREMENT,
    group_id VARCHAR(255) NOT NULL,
    route_id VARCHAR(255) NULL,
    parent_id INT NULL,
    tags VARCHAR(255) NULL,
    excludes TEXT NULL,
    ips VARCHAR(255) NULL,
    name VARCHAR(255) NOT NULL,
    rate VARCHAR(11) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port VARCHAR(11) NOT NULL,
    server_port INT NOT NULL,
    cipher VARCHAR(255) NOT NULL,
    obfs CHAR(11) NULL,
    obfs_settings VARCHAR(255) NULL,
    `show` TINYINT NOT NULL DEFAULT 0,
    sort INT NULL,
    created_at INT NOT NULL,
    updated_at INT NOT NULL,
    
    PRIMARY KEY (id),
    INDEX idx_shadowsocks_servers_group (group_id),
    INDEX idx_shadowsocks_servers_name (name),
    INDEX idx_shadowsocks_servers_host (host),
    INDEX idx_shadowsocks_servers_show (`show`),
    INDEX idx_shadowsocks_servers_sort (sort)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Shadowsocks server configurations';

-- User subscription server group associations
CREATE TABLE user_subscription_server_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    server_group_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    INDEX idx_user_subscription_server_groups_subscription (user_subscription_id),
    INDEX idx_user_subscription_server_groups_group (server_group_id),
    UNIQUE KEY uk_user_subscription_server_groups (user_subscription_id, server_group_id)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Many-to-many relationship between subscriptions and server groups';

-- ==============================================================================
-- SUPPORT SYSTEM
-- ==============================================================================

-- Support tickets
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
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_tickets_ticket_no (ticket_no),
    INDEX idx_tickets_user (user_id),
    INDEX idx_tickets_status (status),
    INDEX idx_tickets_category (category),
    INDEX idx_tickets_priority (priority),
    INDEX idx_tickets_assigned (assigned_to_id),
    INDEX idx_tickets_created_at (created_at),
    INDEX idx_tickets_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Support ticket management system';

-- Ticket messages
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
    INDEX idx_ticket_messages_ticket (ticket_id, created_at),
    INDEX idx_ticket_messages_user (user_id),
    INDEX idx_ticket_messages_type (message_type),
    INDEX idx_ticket_messages_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Ticket message threads';

-- ==============================================================================
-- TRAFFIC MANAGEMENT SYSTEM
-- ==============================================================================

-- Node data for UniProxy integration
CREATE TABLE node_data (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    node_id INT UNSIGNED NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    token VARCHAR(255) NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    content_type VARCHAR(100) NULL,
    user_agent VARCHAR(255) NULL,
    request_body LONGTEXT NOT NULL,
    processed_at TIMESTAMP NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Performance indexes
    INDEX idx_node_data_node_id (node_id),
    INDEX idx_node_data_node_type (node_type),
    INDEX idx_node_data_token (token),
    INDEX idx_node_data_status (status),
    INDEX idx_node_data_processed_at (processed_at),
    INDEX idx_node_data_created_at (created_at),
    INDEX idx_node_data_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Node data for UniProxy traffic management';

-- User traffic logs for detailed tracking
CREATE TABLE user_traffic_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL,
    node_id INT UNSIGNED NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    upload_bytes BIGINT NOT NULL DEFAULT 0,
    download_bytes BIGINT NOT NULL DEFAULT 0,
    total_bytes BIGINT NOT NULL DEFAULT 0,
    node_data_id BIGINT UNSIGNED NOT NULL,
    recorded_at TIMESTAMP NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- High-performance indexes for traffic queries
    INDEX idx_user_traffic_logs_user (user_id),
    INDEX idx_user_traffic_logs_node (node_id),
    INDEX idx_user_traffic_logs_node_type (node_type),
    INDEX idx_user_traffic_logs_recorded_at (recorded_at),
    INDEX idx_user_traffic_logs_total_bytes (total_bytes),
    INDEX idx_user_traffic_logs_node_data (node_data_id),
    INDEX idx_user_traffic_logs_deleted_at (deleted_at),
    
    -- Composite indexes for aggregation
    INDEX idx_user_traffic_logs_user_node_time (user_id, node_id, recorded_at),
    INDEX idx_user_traffic_logs_node_time (node_id, recorded_at),
    INDEX idx_user_traffic_logs_user_time_bytes (user_id, recorded_at, total_bytes)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Detailed user traffic logs for monitoring and billing';

-- ==============================================================================
-- EVENT STORE FOR EVENT-DRIVEN ARCHITECTURE
-- ==============================================================================

-- Event store for comprehensive event sourcing
CREATE TABLE event_store (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    event_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_source VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(100) NULL,
    aggregate_type VARCHAR(50) NULL,
    event_version VARCHAR(20) NOT NULL DEFAULT '1.0',
    event_data TEXT NOT NULL,
    metadata TEXT NULL,
    occurred_at TIMESTAMP NOT NULL,
    stored_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_event_store_event_id (event_id),
    INDEX idx_event_store_event_type (event_type),
    INDEX idx_event_store_event_source (event_source),
    INDEX idx_event_store_aggregate (aggregate_id, aggregate_type),
    INDEX idx_event_store_occurred_at (occurred_at),
    INDEX idx_event_store_stored_at (stored_at),
    
    -- Composite indexes for event replay and querying
    INDEX idx_event_store_type_occurred (event_type, occurred_at),
    INDEX idx_event_store_source_occurred (event_source, occurred_at),
    INDEX idx_event_store_aggregate_occurred (aggregate_id, aggregate_type, occurred_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Event store for event sourcing, audit trail, and domain events';

-- ==============================================================================
-- SYSTEM CONFIGURATION
-- ==============================================================================

-- Settings table for dynamic configuration
CREATE TABLE settings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    key_name VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    description TEXT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'string',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    is_encrypted BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential constraints and indexes
    UNIQUE KEY uk_settings_key_name (key_name),
    INDEX idx_settings_type (type),
    INDEX idx_settings_public (is_public),
    INDEX idx_settings_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='System configuration and dynamic settings';

-- ==============================================================================
-- DATA INITIALIZATION
-- ==============================================================================

-- Initialize default subscription plan
INSERT INTO subscription_plans (
    id, name, code, description, price, currency, billing_cycle, billing_interval, 
    status, is_visible, traffic_limit, traffic_reset_cycle, created_at, updated_at
) VALUES (
    1, 'Basic Plan', 'basic', 'Basic subscription plan with 10GB monthly traffic', 
    9.99, 'USD', 'monthly', 1, 'active', 1, 10737418240, 'monthly', 
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
);

-- Initialize system settings
INSERT INTO settings (key_name, value, description, type, is_public) VALUES
('site_name', 'Linke Service Platform', 'Site name for branding', 'string', TRUE),
('default_currency', 'USD', 'Default system currency', 'string', TRUE),
('payment_timeout_minutes', '30', 'Payment timeout in minutes', 'integer', FALSE),
('invoice_due_days', '30', 'Default invoice due period in days', 'integer', FALSE),
('max_failed_login_attempts', '5', 'Maximum failed login attempts before lockout', 'integer', FALSE),
('account_lockout_duration_minutes', '30', 'Account lockout duration in minutes', 'integer', FALSE),
('jwt_expiration_hours', '24', 'JWT token expiration in hours', 'integer', FALSE),
('traffic_reset_day', '1', 'Day of month for traffic reset (1-28)', 'integer', FALSE),
('max_retry_attempts', '3', 'Maximum payment retry attempts', 'integer', FALSE),
('retry_backoff_multiplier', '2', 'Exponential backoff multiplier for retries', 'integer', FALSE);

-- Initialize default alert configurations (templates)
INSERT INTO alert_configurations (
    user_subscription_id, usage_type, threshold_type, threshold, 
    name, description, priority, notification_channels, cooldown_minutes
) VALUES 
(0, 'traffic', 'percentage', 50.0, 'Traffic 50% Warning', 
 'Alert when traffic usage reaches 50% of limit', 'low', 
 '[{"type":"in_app","target":"user","enabled":true}]', 1440),
(0, 'traffic', 'percentage', 80.0, 'Traffic 80% Alert', 
 'Alert when traffic usage reaches 80% of limit', 'medium', 
 '[{"type":"email","target":"user","enabled":true},{"type":"in_app","target":"user","enabled":true}]', 720),
(0, 'traffic', 'percentage', 90.0, 'Traffic 90% Critical', 
 'Critical alert when traffic usage reaches 90% of limit', 'high', 
 '[{"type":"email","target":"user","enabled":true},{"type":"in_app","target":"user","enabled":true}]', 360),
(0, 'traffic', 'percentage', 100.0, 'Traffic Limit Exceeded', 
 'Critical alert when traffic limit is exceeded', 'critical', 
 '[{"type":"email","target":"user","enabled":true},{"type":"in_app","target":"user","enabled":true}]', 60);