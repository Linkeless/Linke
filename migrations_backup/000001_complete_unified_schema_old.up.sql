-- ==============================================================================
-- COMPLETE UNIFIED SCHEMA MIGRATION
-- ==============================================================================
-- This is a comprehensive migration that creates the complete database schema
-- for the Linke platform. It consolidates all tables from multiple previous
-- migrations into a single, optimized migration following MySQL best practices.
--
-- Architecture: VSA (Vertical Slice Architecture) + Clean Architecture
-- Database: MySQL 8.0+ with utf8mb4 charset
-- Constraints: NO FOREIGN KEY constraints (application-level enforcement)
-- Performance: Optimized indexing strategies for high-traffic scenarios
-- ==============================================================================

-- ==============================================================================
-- CORE SYSTEM TABLES
-- ==============================================================================

-- System configuration and dynamic settings
CREATE TABLE settings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Setting Identity
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT NULL,
    setting_type VARCHAR(20) NOT NULL DEFAULT 'string',
    
    -- Metadata
    description TEXT NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    is_encrypted BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    UNIQUE INDEX idx_settings_key (setting_key),
    INDEX idx_settings_type (setting_type),
    INDEX idx_settings_public (is_public),
    INDEX idx_settings_created_at (created_at),
    INDEX idx_settings_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='System configuration and dynamic settings';

-- ==============================================================================
-- USER MANAGEMENT SYSTEM
-- ==============================================================================

-- Core user management with OAuth support and referral system
CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Core Identity
    email VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NULL,
    
    -- OAuth Integration
    oauth_provider VARCHAR(50) NULL,
    oauth_provider_id VARCHAR(100) NULL,
    oauth_data TEXT NULL,
    
    -- Profile Information
    first_name VARCHAR(100) NULL,
    last_name VARCHAR(100) NULL,
    avatar_url VARCHAR(500) NULL,
    phone VARCHAR(20) NULL,
    
    -- Referral System Integration
    referral_code VARCHAR(20) NULL,
    referred_by_id BIGINT UNSIGNED NULL,
    referral_count INT NOT NULL DEFAULT 0,
    
    -- Subscription References (no FK constraints)
    current_subscription_id BIGINT UNSIGNED NULL,
    subscription_status VARCHAR(20) NOT NULL DEFAULT 'none',
    
    -- User Status and Security
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMP NULL,
    
    -- Security and Access Control
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    permissions TEXT NULL,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45) NULL,
    
    -- Account Lifecycle
    trial_used BOOLEAN NOT NULL DEFAULT FALSE,
    trial_started_at TIMESTAMP NULL,
    trial_ends_at TIMESTAMP NULL,
    
    -- Billing and Localization
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Notification Preferences
    notification_preferences TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_users_email (email),
    UNIQUE INDEX idx_users_username (username),
    UNIQUE INDEX idx_users_referral_code (referral_code),
    
    -- Essential indexes
    INDEX idx_users_oauth (oauth_provider, oauth_provider_id),
    INDEX idx_users_referred_by (referred_by_id),
    INDEX idx_users_subscription (current_subscription_id),
    INDEX idx_users_status (status),
    INDEX idx_users_role (role),
    INDEX idx_users_trial (trial_used, trial_ends_at),
    INDEX idx_users_created_at (created_at),
    INDEX idx_users_deleted_at (deleted_at),
    
    -- Composite indexes for common queries
    INDEX idx_users_active_subscribers (status, subscription_status),
    INDEX idx_users_referral_analysis (referred_by_id, created_at, status),
    INDEX idx_users_trial_conversion (trial_used, subscription_status, created_at),
    
    -- Covering index for user listings
    INDEX idx_users_covering (status, role, email, username, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Core user management with OAuth support and referral system';

-- ==============================================================================
-- AUTHENTICATION AND SECURITY
-- ==============================================================================

-- JWT token blacklist for secure logout
CREATE TABLE jwt_blacklist (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- JWT Token Information
    token_hash VARCHAR(255) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    reason VARCHAR(100) NOT NULL DEFAULT 'logout',
    
    -- Security Context
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    
    -- Unique constraint on token
    UNIQUE INDEX idx_jwt_blacklist_token (token_hash),
    
    -- Essential indexes
    INDEX idx_jwt_blacklist_user (user_id),
    INDEX idx_jwt_blacklist_expires (expires_at),
    INDEX idx_jwt_blacklist_created_at (created_at),
    
    -- Cleanup optimization
    INDEX idx_jwt_blacklist_cleanup (expires_at, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='JWT token blacklist for secure logout';

-- Login attempt tracking for security analysis
CREATE TABLE login_attempts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Attempt Information
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent VARCHAR(500) NULL,
    
    -- Attempt Result
    success BOOLEAN NOT NULL DEFAULT FALSE,
    failure_reason VARCHAR(100) NULL,
    
    -- Security Context
    attempt_count INT NOT NULL DEFAULT 1,
    last_attempt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Rate Limiting
    window_start TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_login_attempts_email (email),
    INDEX idx_login_attempts_ip (ip_address),
    INDEX idx_login_attempts_success (success),
    INDEX idx_login_attempts_last_attempt (last_attempt),
    INDEX idx_login_attempts_window (window_start),
    INDEX idx_login_attempts_created_at (created_at),
    
    -- Composite indexes for rate limiting
    INDEX idx_login_attempts_rate_limit (email, window_start, attempt_count),
    INDEX idx_login_attempts_ip_limit (ip_address, window_start, attempt_count),
    INDEX idx_login_attempts_security (ip_address, email, success, last_attempt)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Login attempt tracking for security analysis';

-- Account lockout management
CREATE TABLE account_lockouts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Lockout Target
    user_id BIGINT UNSIGNED NULL,
    email VARCHAR(255) NULL,
    ip_address VARCHAR(45) NULL,
    
    -- Lockout Details
    lockout_type VARCHAR(20) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    attempt_count INT NOT NULL DEFAULT 0,
    
    -- Timing
    locked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    locked_until TIMESTAMP NOT NULL DEFAULT '2025-01-01 00:00:00',
    unlocked_at TIMESTAMP NULL,
    
    -- Context
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_account_lockouts_user (user_id),
    INDEX idx_account_lockouts_email (email),
    INDEX idx_account_lockouts_ip (ip_address),
    INDEX idx_account_lockouts_type (lockout_type),
    INDEX idx_account_lockouts_locked_until (locked_until),
    INDEX idx_account_lockouts_created_at (created_at),
    
    -- Composite indexes for lockout checks
    INDEX idx_account_lockouts_active_user (user_id, locked_until, unlocked_at),
    INDEX idx_account_lockouts_active_ip (ip_address, locked_until, unlocked_at),
    INDEX idx_account_lockouts_cleanup (locked_until, unlocked_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Account lockout management';

-- ==============================================================================
-- SUBSCRIPTION MANAGEMENT
-- ==============================================================================

-- Subscription plans with traffic and server group configuration
CREATE TABLE subscription_plans (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Plan Identity
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT NULL,
    
    -- Plan Configuration
    plan_type VARCHAR(20) NOT NULL DEFAULT 'standard',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Traffic Allowances
    monthly_traffic_gb BIGINT NOT NULL DEFAULT 0,
    traffic_overage_rate DECIMAL(10,4) NOT NULL DEFAULT 0.0000,
    traffic_reset_day INT NOT NULL DEFAULT 1,
    
    -- Connection Limits
    concurrent_connections INT NOT NULL DEFAULT 10,
    device_limit INT NOT NULL DEFAULT 5,
    
    -- Server Access
    server_group_ids TEXT NULL,
    allowed_protocols TEXT NULL,
    
    -- Pricing
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    
    -- Plan Features
    features TEXT NULL,
    limitations TEXT NULL,
    
    -- Trial Configuration
    trial_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    trial_days INT NOT NULL DEFAULT 0,
    trial_traffic_gb BIGINT NOT NULL DEFAULT 0,
    
    -- Plan Lifecycle
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_subscription_plans_code (code),
    
    -- Essential indexes
    INDEX idx_subscription_plans_name (name),
    INDEX idx_subscription_plans_type (plan_type),
    INDEX idx_subscription_plans_status (status),
    INDEX idx_subscription_plans_public (is_public),
    INDEX idx_subscription_plans_featured (is_featured),
    INDEX idx_subscription_plans_sort (sort_order),
    INDEX idx_subscription_plans_billing (billing_cycle),
    INDEX idx_subscription_plans_trial (trial_enabled),
    INDEX idx_subscription_plans_created_at (created_at),
    INDEX idx_subscription_plans_deleted_at (deleted_at),
    
    -- Composite indexes for plan selection
    INDEX idx_subscription_plans_active (status, is_public, sort_order),
    INDEX idx_subscription_plans_pricing (currency, billing_cycle, price),
    INDEX idx_subscription_plans_trial_plans (trial_enabled, trial_days, status)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Subscription plans with traffic and server group configuration';

-- User subscriptions with comprehensive lifecycle management
CREATE TABLE user_subscriptions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Subscription Identity
    uuid VARCHAR(36) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_plan_id BIGINT UNSIGNED NOT NULL,
    
    -- Subscription Lifecycle
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    start_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMP NULL,
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Trial Configuration
    is_trial BOOLEAN NOT NULL DEFAULT FALSE,
    trial_start_date TIMESTAMP NULL,
    trial_end_date TIMESTAMP NULL,
    trial_converted BOOLEAN NOT NULL DEFAULT FALSE,
    trial_converted_at TIMESTAMP NULL,
    
    -- Traffic Management
    traffic_limit BIGINT NOT NULL DEFAULT 0,
    traffic_used BIGINT NOT NULL DEFAULT 0,
    traffic_reset_date TIMESTAMP NULL,
    traffic_suspended BOOLEAN NOT NULL DEFAULT FALSE,
    traffic_suspended_at TIMESTAMP NULL,
    
    -- Server Access Control
    server_group_ids TEXT NULL,
    allowed_protocols TEXT NULL,
    connection_limit INT NOT NULL DEFAULT 10,
    device_limit INT NOT NULL DEFAULT 5,
    
    -- Subscription Configuration
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Payment Integration
    last_payment_date TIMESTAMP NULL,
    next_payment_date TIMESTAMP NULL,
    payment_method_id BIGINT UNSIGNED NULL,
    
    -- Pause/Resume Management
    paused BOOLEAN NOT NULL DEFAULT FALSE,
    paused_at TIMESTAMP NULL,
    paused_reason VARCHAR(255) NULL,
    pause_days_remaining INT NOT NULL DEFAULT 0,
    
    -- Usage Statistics
    total_traffic_used BIGINT NOT NULL DEFAULT 0,
    peak_concurrent_connections INT NOT NULL DEFAULT 0,
    devices_connected INT NOT NULL DEFAULT 0,
    
    -- Metadata
    metadata TEXT NULL,
    notes TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_user_subscriptions_uuid (uuid),
    
    -- Essential indexes
    INDEX idx_user_subscriptions_user (user_id),
    INDEX idx_user_subscriptions_plan (subscription_plan_id),
    INDEX idx_user_subscriptions_status (status),
    INDEX idx_user_subscriptions_trial (is_trial),
    INDEX idx_user_subscriptions_auto_renew (auto_renew),
    INDEX idx_user_subscriptions_paused (paused),
    INDEX idx_user_subscriptions_traffic_suspended (traffic_suspended),
    INDEX idx_user_subscriptions_start_date (start_date),
    INDEX idx_user_subscriptions_end_date (end_date),
    INDEX idx_user_subscriptions_next_payment (next_payment_date),
    INDEX idx_user_subscriptions_created_at (created_at),
    INDEX idx_user_subscriptions_deleted_at (deleted_at),
    
    -- Composite indexes for common queries
    INDEX idx_user_subscriptions_active (user_id, status, end_date),
    INDEX idx_user_subscriptions_renewal (auto_renew, next_payment_date, status),
    INDEX idx_user_subscriptions_trial_conversion (is_trial, trial_converted, trial_end_date),
    INDEX idx_user_subscriptions_traffic_management (traffic_suspended, traffic_limit, traffic_used),
    INDEX idx_user_subscriptions_billing_cycle (billing_cycle, next_payment_date, status),
    
    -- Covering index for subscription listings
    INDEX idx_user_subscriptions_covering (user_id, status, subscription_plan_id, end_date, traffic_used, traffic_limit)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='User subscriptions with comprehensive lifecycle management';

-- ==============================================================================
-- ORDER AND BILLING MANAGEMENT
-- ==============================================================================

-- Subscription purchase orders with comprehensive billing
CREATE TABLE subscription_orders (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Order Identity
    order_number VARCHAR(50) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_plan_id BIGINT UNSIGNED NOT NULL,
    
    -- Order Details
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    order_type VARCHAR(20) NOT NULL DEFAULT 'new',
    
    -- Pricing Information
    subtotal DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    tax_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Discount/Coupon Integration
    coupon_code VARCHAR(50) NULL,
    coupon_discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    coupon_discount_type VARCHAR(20) NULL,
    
    -- Billing Configuration
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    billing_start_date TIMESTAMP NULL,
    billing_end_date TIMESTAMP NULL,
    
    -- Payment Integration
    payment_method VARCHAR(50) NULL,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_gateway VARCHAR(50) NULL,
    payment_gateway_order_id VARCHAR(100) NULL,
    
    -- Order Fulfillment
    processed_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    cancelled_at TIMESTAMP NULL,
    cancellation_reason VARCHAR(255) NULL,
    
    -- Customer Information (snapshot at time of order)
    customer_email VARCHAR(255) NOT NULL,
    customer_name VARCHAR(200) NULL,
    billing_address TEXT NULL,
    
    -- Invoice Integration
    invoice_id BIGINT UNSIGNED NULL,
    
    -- Metadata
    metadata TEXT NULL,
    notes TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_subscription_orders_number (order_number),
    
    -- Essential indexes
    INDEX idx_subscription_orders_user (user_id),
    INDEX idx_subscription_orders_plan (subscription_plan_id),
    INDEX idx_subscription_orders_status (status),
    INDEX idx_subscription_orders_type (order_type),
    INDEX idx_subscription_orders_payment_status (payment_status),
    INDEX idx_subscription_orders_gateway (payment_gateway),
    INDEX idx_subscription_orders_gateway_order (payment_gateway_order_id),
    INDEX idx_subscription_orders_coupon (coupon_code),
    INDEX idx_subscription_orders_invoice (invoice_id),
    INDEX idx_subscription_orders_email (customer_email),
    INDEX idx_subscription_orders_created_at (created_at),
    INDEX idx_subscription_orders_deleted_at (deleted_at),
    
    -- Composite indexes for reporting and analytics
    INDEX idx_subscription_orders_user_status (user_id, status, created_at),
    INDEX idx_subscription_orders_payment_tracking (payment_status, payment_gateway, created_at),
    INDEX idx_subscription_orders_revenue_analysis (status, total_amount, currency, created_at),
    INDEX idx_subscription_orders_fulfillment (status, processed_at, completed_at),
    
    -- Covering index for order management
    INDEX idx_subscription_orders_covering (user_id, status, order_type, total_amount, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Subscription purchase orders with comprehensive billing';

-- ==============================================================================
-- PAYMENT PROCESSING SYSTEM
-- ==============================================================================

-- Payment records with security and anti-fraud protection
CREATE TABLE payment_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Payment Identity
    payment_no VARCHAR(50) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Order/Subscription Integration
    subscription_order_id BIGINT UNSIGNED NULL,
    user_subscription_id BIGINT UNSIGNED NULL,
    invoice_id BIGINT UNSIGNED NULL,
    
    -- Payment Details
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50) NOT NULL,
    
    -- Gateway Integration
    payment_gateway VARCHAR(50) NOT NULL,
    gateway_transaction_id VARCHAR(100) NULL,
    gateway_order_id VARCHAR(100) NULL,
    gateway_response TEXT NULL,
    
    -- Payment Status and Lifecycle
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMP NULL,
    confirmed_at TIMESTAMP NULL,
    failed_at TIMESTAMP NULL,
    refunded_at TIMESTAMP NULL,
    
    -- Refund Information
    refund_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    refund_reason VARCHAR(255) NULL,
    refunded_by_id BIGINT UNSIGNED NULL,
    
    -- Anti-Fraud and Security
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    risk_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'low',
    
    -- Anti-Replay Protection
    idempotency_key VARCHAR(100) NULL,
    
    -- Callback and Webhook Integration
    callback_url VARCHAR(500) NULL,
    webhook_attempts INT NOT NULL DEFAULT 0,
    webhook_success BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Customer Information (snapshot)
    customer_email VARCHAR(255) NOT NULL,
    customer_name VARCHAR(200) NULL,
    
    -- Metadata
    metadata TEXT NULL,
    notes TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_payment_records_payment_no (payment_no),
    UNIQUE INDEX idx_payment_records_idempotency (idempotency_key),
    
    -- Essential indexes
    INDEX idx_payment_records_user (user_id),
    INDEX idx_payment_records_order (subscription_order_id),
    INDEX idx_payment_records_subscription (user_subscription_id),
    INDEX idx_payment_records_invoice (invoice_id),
    INDEX idx_payment_records_status (status),
    INDEX idx_payment_records_gateway (payment_gateway),
    INDEX idx_payment_records_gateway_txn (gateway_transaction_id),
    INDEX idx_payment_records_method (payment_method),
    INDEX idx_payment_records_paid_at (paid_at),
    INDEX idx_payment_records_risk_level (risk_level),
    INDEX idx_payment_records_email (customer_email),
    INDEX idx_payment_records_created_at (created_at),
    INDEX idx_payment_records_deleted_at (deleted_at),
    
    -- Composite indexes for payment processing
    INDEX idx_payment_records_user_status (user_id, status, created_at),
    INDEX idx_payment_records_gateway_status (payment_gateway, status, created_at),
    INDEX idx_payment_records_processing (status, payment_gateway, created_at),
    INDEX idx_payment_records_fraud_analysis (risk_level, risk_score, status),
    INDEX idx_payment_records_reconciliation (payment_gateway, paid_at, amount),
    
    -- Covering index for payment management
    INDEX idx_payment_records_covering (user_id, status, amount, payment_method, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Payment records with security and anti-fraud protection';

-- Smart payment retry tracking
CREATE TABLE payment_retries (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Retry Identity
    payment_record_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Retry Configuration
    retry_count INT NOT NULL DEFAULT 0,
    max_retry_attempts INT NOT NULL DEFAULT 3,
    retry_strategy VARCHAR(50) NOT NULL DEFAULT 'exponential_backoff',
    
    -- Retry Status and Timing
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    next_retry_at TIMESTAMP NULL,
    last_retry_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    
    -- Failure Analysis
    failure_reasons TEXT NULL,
    permanent_failure BOOLEAN NOT NULL DEFAULT FALSE,
    failure_category VARCHAR(50) NULL,
    
    -- Gateway Context
    payment_gateway VARCHAR(50) NOT NULL,
    original_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Backoff Configuration
    initial_delay_seconds INT NOT NULL DEFAULT 300,
    max_delay_seconds INT NOT NULL DEFAULT 86400,
    backoff_multiplier DECIMAL(3,2) NOT NULL DEFAULT 2.00,
    
    -- Success Tracking
    successful_retry_attempt INT NULL,
    successful_at TIMESTAMP NULL,
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_payment_retries_payment (payment_record_id),
    INDEX idx_payment_retries_user (user_id),
    INDEX idx_payment_retries_status (status),
    INDEX idx_payment_retries_next_retry (next_retry_at),
    INDEX idx_payment_retries_gateway (payment_gateway),
    INDEX idx_payment_retries_failure_category (failure_category),
    INDEX idx_payment_retries_permanent_failure (permanent_failure),
    INDEX idx_payment_retries_created_at (created_at),
    INDEX idx_payment_retries_deleted_at (deleted_at),
    
    -- Composite indexes for retry processing
    INDEX idx_payment_retries_active (status, next_retry_at, retry_count),
    INDEX idx_payment_retries_processing (status, payment_gateway, next_retry_at),
    INDEX idx_payment_retries_analysis (failure_category, permanent_failure, status),
    INDEX idx_payment_retries_success_tracking (successful_at, successful_retry_attempt),
    
    -- Covering index for retry management
    INDEX idx_payment_retries_covering (payment_record_id, status, retry_count, next_retry_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Smart payment retry tracking';

-- Payment retry attempt audit trail
CREATE TABLE payment_retry_histories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- History Identity
    payment_retry_id BIGINT UNSIGNED NOT NULL,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    attempt_number INT NOT NULL,
    
    -- Attempt Details
    attempted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL,
    
    -- Gateway Response
    gateway_response TEXT NULL,
    gateway_error_code VARCHAR(50) NULL,
    gateway_error_message TEXT NULL,
    
    -- Technical Details
    processing_time_ms INT NULL,
    http_status_code INT NULL,
    
    -- Failure Analysis
    failure_reason VARCHAR(255) NULL,
    failure_category VARCHAR(50) NULL,
    is_retryable BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Next Retry Calculation
    next_retry_delay_seconds INT NULL,
    next_retry_scheduled_at TIMESTAMP NULL,
    
    -- Context
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_payment_retry_histories_retry (payment_retry_id),
    INDEX idx_payment_retry_histories_payment (payment_record_id),
    INDEX idx_payment_retry_histories_attempt (attempt_number),
    INDEX idx_payment_retry_histories_status (status),
    INDEX idx_payment_retry_histories_attempted_at (attempted_at),
    INDEX idx_payment_retry_histories_failure_category (failure_category),
    INDEX idx_payment_retry_histories_retryable (is_retryable),
    INDEX idx_payment_retry_histories_created_at (created_at),
    
    -- Composite indexes for retry analysis
    INDEX idx_payment_retry_histories_analysis (payment_retry_id, attempt_number, status),
    INDEX idx_payment_retry_histories_failure_analysis (failure_category, is_retryable, status),
    INDEX idx_payment_retry_histories_performance (processing_time_ms, http_status_code, attempted_at),
    
    -- Covering index for retry tracking
    INDEX idx_payment_retry_histories_covering (payment_retry_id, attempt_number, status, attempted_at, failure_reason)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Payment retry attempt audit trail';

-- Tokenized payment methods for PCI compliance
CREATE TABLE payment_methods (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Method Identity
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Payment Method Details
    payment_type VARCHAR(50) NOT NULL,
    gateway VARCHAR(50) NOT NULL,
    
    -- Tokenized Information (PCI Compliant)
    gateway_token VARCHAR(255) NOT NULL,
    gateway_customer_id VARCHAR(100) NULL,
    
    -- Display Information (Safe for display)
    display_name VARCHAR(100) NOT NULL,
    last_four VARCHAR(4) NULL,
    brand VARCHAR(50) NULL,
    
    -- Card/Method Details
    expiry_month INT NULL,
    expiry_year INT NULL,
    card_type VARCHAR(20) NULL,
    
    -- Billing Information
    billing_name VARCHAR(200) NULL,
    billing_email VARCHAR(255) NULL,
    billing_address TEXT NULL,
    
    -- Method Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMP NULL,
    
    -- Usage Statistics
    usage_count INT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP NULL,
    
    -- Security
    fingerprint VARCHAR(255) NULL,
    risk_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_payment_methods_user (user_id),
    INDEX idx_payment_methods_gateway (gateway),
    INDEX idx_payment_methods_token (gateway_token),
    INDEX idx_payment_methods_customer (gateway_customer_id),
    INDEX idx_payment_methods_type (payment_type),
    INDEX idx_payment_methods_status (status),
    INDEX idx_payment_methods_default (is_default),
    INDEX idx_payment_methods_verified (verified),
    INDEX idx_payment_methods_fingerprint (fingerprint),
    INDEX idx_payment_methods_created_at (created_at),
    INDEX idx_payment_methods_deleted_at (deleted_at),
    
    -- Composite indexes for payment method management
    INDEX idx_payment_methods_user_active (user_id, status, is_default),
    INDEX idx_payment_methods_user_default (user_id, is_default, status),
    INDEX idx_payment_methods_gateway_management (gateway, status, verified),
    INDEX idx_payment_methods_usage_tracking (usage_count, last_used_at, status),
    
    -- Covering index for method selection
    INDEX idx_payment_methods_covering (user_id, status, payment_type, display_name, is_default)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Tokenized payment methods for PCI compliance';

-- Payment gateway configurations
CREATE TABLE payment_configs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Configuration Identity
    gateway VARCHAR(50) NOT NULL,
    environment VARCHAR(20) NOT NULL DEFAULT 'production',
    
    -- Gateway Configuration
    gateway_name VARCHAR(100) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- API Configuration (encrypted)
    api_endpoint VARCHAR(500) NULL,
    api_key VARCHAR(255) NULL,
    api_secret VARCHAR(255) NULL,
    merchant_id VARCHAR(100) NULL,
    
    -- Payment Configuration
    supported_currencies TEXT NULL,
    supported_payment_methods TEXT NULL,
    min_amount DECIMAL(10,2) NOT NULL DEFAULT 0.01,
    max_amount DECIMAL(10,2) NOT NULL DEFAULT 999999.99,
    
    -- Features
    supports_refunds BOOLEAN NOT NULL DEFAULT TRUE,
    supports_partial_refunds BOOLEAN NOT NULL DEFAULT TRUE,
    supports_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    supports_webhooks BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Webhook Configuration
    webhook_url VARCHAR(500) NULL,
    webhook_secret VARCHAR(255) NULL,
    webhook_events TEXT NULL,
    
    -- Processing Configuration
    processing_delay_seconds INT NOT NULL DEFAULT 0,
    auto_capture BOOLEAN NOT NULL DEFAULT TRUE,
    require_cvv BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Rate Limiting
    rate_limit_requests_per_minute INT NOT NULL DEFAULT 60,
    rate_limit_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_payment_configs_gateway_env (gateway, environment),
    
    -- Essential indexes
    INDEX idx_payment_configs_gateway (gateway),
    INDEX idx_payment_configs_environment (environment),
    INDEX idx_payment_configs_enabled (is_enabled),
    INDEX idx_payment_configs_recurring (supports_recurring),
    INDEX idx_payment_configs_webhooks (supports_webhooks),
    INDEX idx_payment_configs_created_at (created_at),
    INDEX idx_payment_configs_deleted_at (deleted_at),
    
    -- Composite indexes for gateway selection
    INDEX idx_payment_configs_active (is_enabled, environment, gateway),
    INDEX idx_payment_configs_features (supports_refunds, supports_recurring, is_enabled)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Payment gateway configurations';

-- ==============================================================================
-- INVOICE MANAGEMENT
-- ==============================================================================

-- Comprehensive invoice management system
CREATE TABLE invoices (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Invoice Identity
    invoice_number VARCHAR(50) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_order_id BIGINT UNSIGNED NULL,
    
    -- Invoice Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    
    -- Billing Information
    subtotal DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Tax Details
    tax_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    tax_description VARCHAR(100) NULL,
    
    -- Customer Information (snapshot)
    customer_name VARCHAR(200) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    billing_address TEXT NULL,
    
    -- Payment Information
    payment_terms VARCHAR(50) NOT NULL DEFAULT 'immediate',
    payment_due_date TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,
    payment_method VARCHAR(50) NULL,
    
    -- Invoice Dates
    issue_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    service_period_start TIMESTAMP NULL,
    service_period_end TIMESTAMP NULL,
    
    -- PDF Generation and Storage
    pdf_generated BOOLEAN NOT NULL DEFAULT FALSE,
    pdf_file_path VARCHAR(500) NULL,
    pdf_generated_at TIMESTAMP NULL,
    pdf_size_bytes BIGINT NULL,
    
    -- Download Tracking and Security
    download_count INT NOT NULL DEFAULT 0,
    last_downloaded_at TIMESTAMP NULL,
    download_token VARCHAR(100) NULL,
    download_expires_at TIMESTAMP NULL,
    
    -- Email Tracking
    email_sent BOOLEAN NOT NULL DEFAULT FALSE,
    email_sent_at TIMESTAMP NULL,
    
    -- Line Items (JSON for flexibility)
    line_items TEXT NULL,
    
    -- Metadata
    metadata TEXT NULL,
    notes TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_invoices_number (invoice_number),
    
    -- Essential indexes
    INDEX idx_invoices_user (user_id),
    INDEX idx_invoices_order (subscription_order_id),
    INDEX idx_invoices_status (status),
    INDEX idx_invoices_issue_date (issue_date),
    INDEX idx_invoices_due_date (payment_due_date),
    INDEX idx_invoices_paid_at (paid_at),
    INDEX idx_invoices_email (customer_email),
    INDEX idx_invoices_pdf_generated (pdf_generated),
    INDEX idx_invoices_download_token (download_token),
    INDEX idx_invoices_created_at (created_at),
    INDEX idx_invoices_deleted_at (deleted_at),
    
    -- Composite indexes for invoice management
    INDEX idx_invoices_user_status (user_id, status, issue_date),
    INDEX idx_invoices_payment_tracking (status, payment_due_date, paid_at),
    INDEX idx_invoices_pdf_management (pdf_generated, pdf_generated_at, status),
    INDEX idx_invoices_download_security (download_token, download_expires_at),
    
    -- Covering index for invoice listings
    INDEX idx_invoices_covering (user_id, status, total_amount, issue_date, paid_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Comprehensive invoice management system';

-- ==============================================================================
-- USAGE TRACKING AND MONITORING
-- ==============================================================================

-- High-performance usage tracking for real-time monitoring
CREATE TABLE usage_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Usage Identity
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
    
    -- Essential indexes for high-performance queries
    INDEX idx_usage_records_subscription (user_subscription_id),
    INDEX idx_usage_records_usage_type (usage_type),
    INDEX idx_usage_records_timestamp (timestamp),
    INDEX idx_usage_records_source_type (source_type),
    INDEX idx_usage_records_source_id (source_id),
    INDEX idx_usage_records_ip_address (ip_address),
    INDEX idx_usage_records_created_at (created_at),
    INDEX idx_usage_records_deleted_at (deleted_at),
    
    -- Time-series optimized composite indexes
    INDEX idx_usage_records_subscription_type_time (user_subscription_id, usage_type, timestamp),
    INDEX idx_usage_records_type_time_amount (usage_type, timestamp, amount),
    INDEX idx_usage_records_source_analytics (source_type, source_id, timestamp),
    INDEX idx_usage_records_daily_aggregation (user_subscription_id, usage_type, created_at),
    INDEX idx_usage_records_hourly_tracking (user_subscription_id, usage_type, timestamp, created_at),
    
    -- Covering index for usage summaries (most common query pattern)
    INDEX idx_usage_records_covering (user_subscription_id, usage_type, timestamp, amount, unit),
    
    -- Partitioning-ready indexes for very large datasets
    INDEX idx_usage_records_partition_ready (created_at, user_subscription_id, usage_type)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='High-performance usage tracking for real-time monitoring';

-- User-configured alert thresholds for usage monitoring
CREATE TABLE alert_configurations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Configuration Identity
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    
    -- Alert Settings
    usage_type VARCHAR(50) NOT NULL COMMENT 'Type of usage to monitor',
    threshold_type VARCHAR(20) NOT NULL DEFAULT 'percentage' COMMENT 'percentage or absolute',
    threshold DECIMAL(10,2) NOT NULL COMMENT 'Alert threshold value',
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
    INDEX idx_alert_configs_threshold_type (threshold_type),
    INDEX idx_alert_configs_created_at (created_at),
    INDEX idx_alert_configs_deleted_at (deleted_at),
    
    -- Composite indexes for alert processing
    INDEX idx_alert_configs_active (user_subscription_id, is_enabled, deleted_at),
    INDEX idx_alert_configs_monitoring (usage_type, is_enabled, threshold_type),
    INDEX idx_alert_configs_priority_enabled (priority, is_enabled, created_at),
    
    -- Covering index for alert configuration queries
    INDEX idx_alert_configs_covering (user_subscription_id, usage_type, is_enabled, threshold, threshold_type, priority)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='User-configured alert thresholds for usage monitoring';

-- Fired usage alerts and notification tracking
CREATE TABLE usage_alerts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Alert Identity
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    alert_configuration_id BIGINT UNSIGNED NOT NULL,
    
    -- Alert Details
    usage_type VARCHAR(50) NOT NULL,
    current_usage BIGINT NOT NULL COMMENT 'Usage when alert fired',
    usage_limit BIGINT NOT NULL COMMENT 'Usage limit at time of alert',
    threshold_value DECIMAL(10,2) NOT NULL COMMENT 'Threshold that was exceeded',
    usage_percent DECIMAL(5,2) NOT NULL COMMENT 'Percentage of limit used',
    
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
    INDEX idx_usage_alerts_last_notification (last_notification_sent),
    INDEX idx_usage_alerts_created_at (created_at),
    INDEX idx_usage_alerts_deleted_at (deleted_at),
    
    -- Composite indexes for alert management
    INDEX idx_usage_alerts_active (user_subscription_id, status, resolved_at),
    INDEX idx_usage_alerts_monitoring (alert_configuration_id, status, fired_at),
    INDEX idx_usage_alerts_severity_status (severity, status, fired_at),
    INDEX idx_usage_alerts_notification_tracking (notifications_sent, last_notification_sent, status),
    INDEX idx_usage_alerts_resolution_tracking (status, fired_at, resolved_at),
    
    -- Covering index for alert queries
    INDEX idx_usage_alerts_covering (user_subscription_id, status, severity, fired_at, usage_percent, threshold_value)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Fired usage alerts and notification tracking';

-- ==============================================================================
-- COUPON SYSTEM
-- ==============================================================================

-- Comprehensive coupon management system
CREATE TABLE coupons (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Coupon Identity
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    
    -- Coupon Type and Discount
    discount_type VARCHAR(20) NOT NULL,
    discount_value DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    
    -- Usage Limits
    max_uses INT NULL,
    used_count INT NOT NULL DEFAULT 0,
    max_uses_per_user INT NULL,
    
    -- Validity Period
    starts_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    
    -- Applicability Rules
    min_order_amount DECIMAL(10,2) NULL,
    applicable_plans TEXT NULL,
    applicable_user_types TEXT NULL,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Metadata
    metadata TEXT NULL,
    created_by_id BIGINT UNSIGNED NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_coupons_code (code),
    
    -- Essential indexes
    INDEX idx_coupons_name (name),
    INDEX idx_coupons_discount_type (discount_type),
    INDEX idx_coupons_status (status),
    INDEX idx_coupons_public (is_public),
    INDEX idx_coupons_starts_at (starts_at),
    INDEX idx_coupons_expires_at (expires_at),
    INDEX idx_coupons_created_by (created_by_id),
    INDEX idx_coupons_created_at (created_at),
    INDEX idx_coupons_deleted_at (deleted_at),
    
    -- Composite indexes for coupon validation
    INDEX idx_coupons_active (status, starts_at, expires_at),
    INDEX idx_coupons_usage_tracking (max_uses, used_count, status),
    INDEX idx_coupons_public_active (is_public, status, expires_at),
    
    -- Covering index for coupon selection
    INDEX idx_coupons_covering (code, status, discount_type, discount_value, expires_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Comprehensive coupon management system';

-- Coupon usage tracking and fraud prevention
CREATE TABLE coupon_usages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Usage Identity
    coupon_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    subscription_order_id BIGINT UNSIGNED NULL,
    
    -- Usage Details
    discount_amount DECIMAL(10,2) NOT NULL,
    order_amount DECIMAL(10,2) NOT NULL,
    
    -- Context
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_coupon_usages_coupon (coupon_id),
    INDEX idx_coupon_usages_user (user_id),
    INDEX idx_coupon_usages_order (subscription_order_id),
    INDEX idx_coupon_usages_created_at (created_at),
    
    -- Composite indexes for usage validation
    INDEX idx_coupon_usages_user_coupon (user_id, coupon_id),
    INDEX idx_coupon_usages_coupon_usage (coupon_id, created_at),
    
    -- Covering index for usage tracking
    INDEX idx_coupon_usages_covering (coupon_id, user_id, discount_amount, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Coupon usage tracking and fraud prevention';

-- ==============================================================================
-- COMPLETE REFERRAL SYSTEM
-- ==============================================================================

-- Invite codes for referral system
CREATE TABLE invite_codes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Core Fields
    code VARCHAR(32) NOT NULL,
    created_by_id BIGINT UNSIGNED NOT NULL,
    
    -- Referral Integration
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

-- Invite code usage audit trail
CREATE TABLE invite_code_usages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Usage Identity
    invite_code_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Usage Context
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    referrer_url VARCHAR(500) NULL,
    
    -- UTM Tracking
    utm_source VARCHAR(100) NULL,
    utm_campaign VARCHAR(100) NULL,
    utm_medium VARCHAR(100) NULL,
    utm_term VARCHAR(100) NULL,
    utm_content VARCHAR(100) NULL,
    
    -- Conversion Tracking
    converted BOOLEAN NOT NULL DEFAULT FALSE,
    converted_at TIMESTAMP NULL,
    conversion_value DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_invite_code_usages_code (invite_code_id),
    INDEX idx_invite_code_usages_user (user_id),
    INDEX idx_invite_code_usages_converted (converted),
    INDEX idx_invite_code_usages_utm_source (utm_source),
    INDEX idx_invite_code_usages_created_at (created_at),
    INDEX idx_invite_code_usages_deleted_at (deleted_at),
    
    -- Composite indexes for analytics
    INDEX idx_invite_code_usages_conversion (invite_code_id, converted, converted_at),
    INDEX idx_invite_code_usages_utm_analysis (utm_source, utm_campaign, utm_medium),
    
    -- Covering index for usage tracking
    INDEX idx_invite_code_usages_covering (invite_code_id, user_id, converted, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Invite code usage audit trail';

-- Referral campaigns for reward management
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
COMMENT='Referral campaigns for reward management';

-- Referral tracking and reward management
CREATE TABLE referrals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Core Fields
    referrer_id BIGINT UNSIGNED NOT NULL,
    referee_id BIGINT UNSIGNED NOT NULL,
    
    -- Referral Source
    invite_code_id BIGINT UNSIGNED NULL,
    referral_source VARCHAR(50) NOT NULL,
    referral_channel VARCHAR(50) NULL,
    referral_code VARCHAR(100) NULL,
    campaign_id BIGINT UNSIGNED NULL,
    
    -- Status and Tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    referee_status VARCHAR(20) NOT NULL DEFAULT 'registered',
    
    -- Conversion Tracking
    converted_at TIMESTAMP NULL,
    conversion_value DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    conversion_type VARCHAR(50) NULL,
    
    -- Reward Tracking
    reward_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    reward_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    referee_reward DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    rewarded_at TIMESTAMP NULL,
    
    -- Attribution Data
    first_click_at TIMESTAMP NULL,
    last_click_at TIMESTAMP NULL,
    click_count INT NOT NULL DEFAULT 0,
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    referrer_url VARCHAR(500) NULL,
    landing_page VARCHAR(500) NULL,
    utm_source VARCHAR(100) NULL,
    utm_campaign VARCHAR(100) NULL,
    utm_medium VARCHAR(100) NULL,
    utm_term VARCHAR(100) NULL,
    utm_content VARCHAR(100) NULL,
    
    -- Expiration
    expires_at TIMESTAMP NULL,
    
    -- Metadata
    metadata TEXT NULL,
    notes TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_referrals_referrer (referrer_id),
    INDEX idx_referrals_referee (referee_id),
    INDEX idx_referrals_invite_code (invite_code_id),
    INDEX idx_referrals_referral_source (referral_source),
    INDEX idx_referrals_campaign (campaign_id),
    INDEX idx_referrals_status (status),
    INDEX idx_referrals_reward_status (reward_status),
    INDEX idx_referrals_created_at (created_at),
    INDEX idx_referrals_deleted_at (deleted_at),
    
    -- Composite indexes for performance
    INDEX idx_referrals_referrer_status (referrer_id, status),
    INDEX idx_referrals_conversion_tracking (status, converted_at),
    INDEX idx_referrals_reward_tracking (referrer_id, reward_status)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Referral tracking and reward management';

-- Referral rewards tracking and payout management
CREATE TABLE referral_rewards (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Links (no FK constraints)
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
    INDEX idx_referral_rewards_conversion_type (conversion_type),
    INDEX idx_referral_rewards_payout_batch (payout_batch_id),
    INDEX idx_referral_rewards_approved_at (approved_at),
    INDEX idx_referral_rewards_created_at (created_at),
    INDEX idx_referral_rewards_deleted_at (deleted_at),
    
    -- Composite indexes for reward management
    INDEX idx_referral_rewards_user_status (user_id, status),
    INDEX idx_referral_rewards_referral_status (referral_id, status),
    INDEX idx_referral_rewards_status_earned (status, earned_at),
    INDEX idx_referral_rewards_approval_workflow (requires_approval, approved_at, status),
    INDEX idx_referral_rewards_payout_analysis (payout_batch_id, status, paid_at),
    
    -- Covering index for reward summaries
    INDEX idx_referral_rewards_covering (user_id, status, reward_type, reward_amount, earned_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Referral rewards tracking and payout management';

-- Referral event tracking for attribution and analytics
CREATE TABLE referral_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Links (no FK constraints)
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
    INDEX idx_referral_events_processed_at (processed_at),
    INDEX idx_referral_events_created_at (created_at),
    INDEX idx_referral_events_deleted_at (deleted_at),
    
    -- Composite indexes for analytics
    INDEX idx_referral_events_referral_type_time (referral_id, event_type, created_at),
    INDEX idx_referral_events_user_type_time (user_id, event_type, created_at),
    INDEX idx_referral_events_utm_analysis (utm_source, utm_campaign, utm_medium),
    INDEX idx_referral_events_conversion_tracking (event_type, event_value, processed_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Referral event tracking for attribution and analytics';

-- ==============================================================================
-- SERVER MANAGEMENT
-- ==============================================================================

-- Server groups for organization and access control
CREATE TABLE server_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Group Identity
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    
    -- Group Configuration
    group_type VARCHAR(50) NOT NULL DEFAULT 'standard',
    region VARCHAR(50) NULL,
    country_code VARCHAR(2) NULL,
    
    -- Access Control
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    requires_premium BOOLEAN NOT NULL DEFAULT FALSE,
    min_subscription_tier VARCHAR(20) NULL,
    
    -- Load Balancing
    load_balancing_method VARCHAR(20) NOT NULL DEFAULT 'round_robin',
    health_check_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_server_groups_name (name),
    INDEX idx_server_groups_type (group_type),
    INDEX idx_server_groups_region (region),
    INDEX idx_server_groups_country (country_code),
    INDEX idx_server_groups_public (is_public),
    INDEX idx_server_groups_premium (requires_premium),
    INDEX idx_server_groups_status (status),
    INDEX idx_server_groups_created_at (created_at),
    INDEX idx_server_groups_deleted_at (deleted_at),
    
    -- Composite indexes for server selection
    INDEX idx_server_groups_active (status, is_public, requires_premium),
    INDEX idx_server_groups_region_active (region, status, is_public)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Server groups for organization and access control';

-- Shadowsocks server configurations
CREATE TABLE shadowsocks_servers (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Server Identity
    name VARCHAR(100) NOT NULL,
    group_id BIGINT UNSIGNED NOT NULL,
    
    -- Server Configuration
    host VARCHAR(255) NOT NULL,
    server_port INT NOT NULL,
    cipher VARCHAR(50) NOT NULL DEFAULT 'aes-256-gcm',
    
    -- Obfuscation (optional)
    obfs VARCHAR(50) NULL,
    obfs_settings VARCHAR(255) NULL,
    
    -- Server Status and Health
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    health_status VARCHAR(20) NOT NULL DEFAULT 'unknown',
    last_health_check TIMESTAMP NULL,
    
    -- Performance Metrics
    current_connections INT NOT NULL DEFAULT 0,
    max_connections INT NOT NULL DEFAULT 1000,
    bandwidth_limit_mbps INT NULL,
    
    -- Location Information
    region VARCHAR(50) NULL,
    country_code VARCHAR(2) NULL,
    city VARCHAR(100) NULL,
    latitude DECIMAL(10,8) NULL,
    longitude DECIMAL(11,8) NULL,
    
    -- Load Balancing
    weight INT NOT NULL DEFAULT 100,
    priority INT NOT NULL DEFAULT 0,
    
    -- Maintenance
    maintenance_mode BOOLEAN NOT NULL DEFAULT FALSE,
    maintenance_message TEXT NULL,
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_shadowsocks_servers_name (name),
    INDEX idx_shadowsocks_servers_group (group_id),
    INDEX idx_shadowsocks_servers_host (host),
    INDEX idx_shadowsocks_servers_status (status),
    INDEX idx_shadowsocks_servers_visible (is_visible),
    INDEX idx_shadowsocks_servers_health (health_status),
    INDEX idx_shadowsocks_servers_region (region),
    INDEX idx_shadowsocks_servers_country (country_code),
    INDEX idx_shadowsocks_servers_maintenance (maintenance_mode),
    INDEX idx_shadowsocks_servers_created_at (created_at),
    INDEX idx_shadowsocks_servers_deleted_at (deleted_at),
    
    -- Composite indexes for server selection
    INDEX idx_shadowsocks_servers_active (status, is_visible, maintenance_mode),
    INDEX idx_shadowsocks_servers_group_active (group_id, status, is_visible),
    INDEX idx_shadowsocks_servers_load_balancing (group_id, weight, priority),
    INDEX idx_shadowsocks_servers_health_monitoring (health_status, last_health_check),
    
    -- Covering index for server listings
    INDEX idx_shadowsocks_servers_covering (group_id, status, is_visible, name, region, current_connections)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Shadowsocks server configurations';

-- Many-to-many relationship between subscriptions and server groups
CREATE TABLE user_subscription_server_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Relationship
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    server_group_id BIGINT UNSIGNED NOT NULL,
    
    -- Access Configuration
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    granted_by_id BIGINT UNSIGNED NULL,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraint
    UNIQUE INDEX idx_user_sub_server_groups_unique (user_subscription_id, server_group_id, deleted_at),
    
    -- Essential indexes
    INDEX idx_user_sub_server_groups_subscription (user_subscription_id),
    INDEX idx_user_sub_server_groups_group (server_group_id),
    INDEX idx_user_sub_server_groups_status (status),
    INDEX idx_user_sub_server_groups_expires (expires_at),
    INDEX idx_user_sub_server_groups_granted_by (granted_by_id),
    INDEX idx_user_sub_server_groups_created_at (created_at),
    INDEX idx_user_sub_server_groups_deleted_at (deleted_at),
    
    -- Composite indexes for access control
    INDEX idx_user_sub_server_groups_access (user_subscription_id, status, expires_at),
    INDEX idx_user_sub_server_groups_group_access (server_group_id, status, expires_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Many-to-many relationship between subscriptions and server groups';

-- ==============================================================================
-- SUPPORT SYSTEM
-- ==============================================================================

-- Support ticket management system
CREATE TABLE tickets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Ticket Identity
    ticket_number VARCHAR(20) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Ticket Details
    subject VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    
    -- Status and Assignment
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    assigned_to_id BIGINT UNSIGNED NULL,
    
    -- Resolution
    resolved_at TIMESTAMP NULL,
    resolution_summary TEXT NULL,
    
    -- Customer Satisfaction
    satisfaction_rating INT NULL,
    satisfaction_feedback TEXT NULL,
    
    -- Metadata
    metadata TEXT NULL,
    tags TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_tickets_number (ticket_number),
    
    -- Essential indexes
    INDEX idx_tickets_user (user_id),
    INDEX idx_tickets_category (category),
    INDEX idx_tickets_priority (priority),
    INDEX idx_tickets_status (status),
    INDEX idx_tickets_assigned_to (assigned_to_id),
    INDEX idx_tickets_resolved_at (resolved_at),
    INDEX idx_tickets_satisfaction (satisfaction_rating),
    INDEX idx_tickets_created_at (created_at),
    INDEX idx_tickets_deleted_at (deleted_at),
    
    -- Composite indexes for ticket management
    INDEX idx_tickets_user_status (user_id, status, created_at),
    INDEX idx_tickets_assigned_status (assigned_to_id, status, priority),
    INDEX idx_tickets_category_priority (category, priority, status),
    
    -- Covering index for ticket listings
    INDEX idx_tickets_covering (status, priority, category, subject, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Support ticket management system';

-- Ticket message threads
CREATE TABLE ticket_messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Message Identity
    ticket_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    
    -- Message Details
    message TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'reply',
    is_internal BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Attachments
    has_attachments BOOLEAN NOT NULL DEFAULT FALSE,
    attachment_count INT NOT NULL DEFAULT 0,
    attachments TEXT NULL,
    
    -- Metadata
    metadata TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_ticket_messages_ticket (ticket_id),
    INDEX idx_ticket_messages_user (user_id),
    INDEX idx_ticket_messages_type (message_type),
    INDEX idx_ticket_messages_internal (is_internal),
    INDEX idx_ticket_messages_attachments (has_attachments),
    INDEX idx_ticket_messages_created_at (created_at),
    INDEX idx_ticket_messages_deleted_at (deleted_at),
    
    -- Composite indexes for message threading
    INDEX idx_ticket_messages_thread (ticket_id, created_at, is_internal),
    INDEX idx_ticket_messages_user_thread (user_id, ticket_id, created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Ticket message threads';

-- ==============================================================================
-- TRAFFIC MANAGEMENT SYSTEM
-- ==============================================================================

-- Node data for UniProxy traffic management
CREATE TABLE node_data (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Node Identity
    node_id INT NOT NULL,
    node_type VARCHAR(50) NOT NULL DEFAULT 'shadowsocks',
    
    -- Node Information
    node_name VARCHAR(100) NULL,
    node_ip VARCHAR(45) NOT NULL,
    node_port INT NOT NULL,
    
    -- Status and Health
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_heartbeat TIMESTAMP NULL,
    online BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Performance Metrics
    cpu_usage DECIMAL(5,2) NULL,
    memory_usage DECIMAL(5,2) NULL,
    disk_usage DECIMAL(5,2) NULL,
    network_usage BIGINT NULL,
    
    -- Connection Statistics
    current_connections INT NOT NULL DEFAULT 0,
    max_connections INT NOT NULL DEFAULT 1000,
    total_connections BIGINT NOT NULL DEFAULT 0,
    
    -- Traffic Statistics
    bytes_uploaded BIGINT NOT NULL DEFAULT 0,
    bytes_downloaded BIGINT NOT NULL DEFAULT 0,
    
    -- Configuration
    config_data TEXT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_node_data_node_type (node_id, node_type),
    
    -- Essential indexes
    INDEX idx_node_data_node_id (node_id),
    INDEX idx_node_data_node_type_only (node_type),
    INDEX idx_node_data_status (status),
    INDEX idx_node_data_online (online),
    INDEX idx_node_data_last_heartbeat (last_heartbeat),
    INDEX idx_node_data_created_at (created_at),
    INDEX idx_node_data_deleted_at (deleted_at),
    
    -- Composite indexes for monitoring
    INDEX idx_node_data_health (status, online, last_heartbeat),
    INDEX idx_node_data_performance (cpu_usage, memory_usage, current_connections),
    
    -- Covering index for node monitoring
    INDEX idx_node_data_covering (node_id, status, online, current_connections, last_heartbeat)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Node data for UniProxy traffic management';

-- Detailed user traffic logs for monitoring and billing
CREATE TABLE user_traffic_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- User Identity
    user_id BIGINT UNSIGNED NOT NULL,
    user_subscription_id BIGINT UNSIGNED NULL,
    
    -- Traffic Details
    bytes_uploaded BIGINT NOT NULL DEFAULT 0,
    bytes_downloaded BIGINT NOT NULL DEFAULT 0,
    total_bytes BIGINT NOT NULL DEFAULT 0,
    
    -- Server Information
    server_id INT NULL,
    server_group_id BIGINT UNSIGNED NULL,
    server_ip VARCHAR(45) NULL,
    
    -- Session Information
    session_id VARCHAR(100) NULL,
    session_start TIMESTAMP NULL,
    session_end TIMESTAMP NULL,
    session_duration INT NULL,
    
    -- Connection Details
    client_ip VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    protocol VARCHAR(20) NOT NULL DEFAULT 'shadowsocks',
    
    -- Quality Metrics
    connection_quality VARCHAR(20) NULL,
    average_speed_mbps DECIMAL(10,2) NULL,
    
    -- Billing Period
    billing_period VARCHAR(7) NOT NULL,
    
    -- Timestamps
    logged_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes for high-performance queries
    INDEX idx_user_traffic_logs_user (user_id),
    INDEX idx_user_traffic_logs_subscription (user_subscription_id),
    INDEX idx_user_traffic_logs_server (server_id),
    INDEX idx_user_traffic_logs_server_group (server_group_id),
    INDEX idx_user_traffic_logs_session (session_id),
    INDEX idx_user_traffic_logs_billing_period (billing_period),
    INDEX idx_user_traffic_logs_logged_at (logged_at),
    INDEX idx_user_traffic_logs_created_at (created_at),
    INDEX idx_user_traffic_logs_deleted_at (deleted_at),
    
    -- Time-series optimized composite indexes
    INDEX idx_user_traffic_logs_user_period (user_id, billing_period, logged_at),
    INDEX idx_user_traffic_logs_subscription_period (user_subscription_id, billing_period, logged_at),
    INDEX idx_user_traffic_logs_server_analytics (server_id, logged_at, total_bytes),
    INDEX idx_user_traffic_logs_session_tracking (session_id, session_start, session_end),
    
    -- Covering index for billing calculations
    INDEX idx_user_traffic_logs_billing_covering (user_id, billing_period, bytes_uploaded, bytes_downloaded, total_bytes, logged_at),
    
    -- Partitioning-ready indexes for very large datasets
    INDEX idx_user_traffic_logs_partition_ready (created_at, user_id, billing_period)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Detailed user traffic logs for monitoring and billing';

-- ==============================================================================
-- EVENT SOURCING
-- ==============================================================================

-- Event store for event sourcing, audit trail, and domain events
CREATE TABLE event_store (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Event Identity
    event_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_version INT NOT NULL DEFAULT 1,
    
    -- Aggregate Information
    aggregate_id VARCHAR(100) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_version INT NOT NULL DEFAULT 1,
    
    -- Event Data
    event_data TEXT NOT NULL,
    metadata TEXT NULL,
    
    -- Causation and Correlation
    causation_id VARCHAR(36) NULL,
    correlation_id VARCHAR(36) NULL,
    
    -- Event Context
    user_id BIGINT UNSIGNED NULL,
    session_id VARCHAR(100) NULL,
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    
    -- Processing Status
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    processed_at TIMESTAMP NULL,
    
    -- Timestamps
    occurred_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    
    -- Unique constraints
    UNIQUE INDEX idx_event_store_event_id (event_id),
    
    -- Essential indexes for event sourcing
    INDEX idx_event_store_event_type (event_type),
    INDEX idx_event_store_aggregate (aggregate_id, aggregate_type),
    INDEX idx_event_store_aggregate_version (aggregate_id, aggregate_version),
    INDEX idx_event_store_user (user_id),
    INDEX idx_event_store_session (session_id),
    INDEX idx_event_store_causation (causation_id),
    INDEX idx_event_store_correlation (correlation_id),
    INDEX idx_event_store_processed (processed),
    INDEX idx_event_store_occurred_at (occurred_at),
    INDEX idx_event_store_created_at (created_at),
    
    -- Composite indexes for event replaying and processing
    INDEX idx_event_store_aggregate_replay (aggregate_id, aggregate_type, aggregate_version, occurred_at),
    INDEX idx_event_store_type_processing (event_type, processed, occurred_at),
    INDEX idx_event_store_causation_chain (causation_id, correlation_id, occurred_at),
    INDEX idx_event_store_user_activity (user_id, event_type, occurred_at),
    
    -- Covering index for event streaming
    INDEX idx_event_store_covering (aggregate_id, event_type, aggregate_version, occurred_at, processed)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Event store for event sourcing, audit trail, and domain events';

-- ==============================================================================
-- DATA INITIALIZATION
-- ==============================================================================

-- Insert default subscription plan
INSERT INTO subscription_plans (
    name, code, description, plan_type, status, 
    monthly_traffic_gb, concurrent_connections, device_limit,
    price, currency, billing_cycle, is_public, is_featured,
    trial_enabled, trial_days, trial_traffic_gb,
    features, metadata
) VALUES (
    'Basic Plan', 'basic', 'Basic subscription plan with essential features', 
    'standard', 'active', 100, 5, 3, 9.99, 'USD', 'monthly', 
    true, false, true, 7, 10,
    '["Unlimited bandwidth", "Basic server access", "Email support"]',
    '{"created_by": "system", "auto_generated": true}'
);

-- Insert essential system settings
INSERT INTO settings (setting_key, setting_value, setting_type, description, is_public) VALUES
('app_name', 'Linke', 'string', 'Application name', true),
('app_version', '1.0.0', 'string', 'Application version', true),
('maintenance_mode', 'false', 'boolean', 'Maintenance mode status', false),
('max_concurrent_sessions', '10', 'integer', 'Maximum concurrent sessions per user', false),
('session_timeout_minutes', '120', 'integer', 'Session timeout in minutes', false),
('password_min_length', '8', 'integer', 'Minimum password length', false),
('jwt_expiry_hours', '24', 'integer', 'JWT token expiry in hours', false),
('rate_limit_requests_per_minute', '100', 'integer', 'Rate limit for API requests', false);

-- Insert default alert configuration templates
INSERT INTO alert_configurations (
    user_subscription_id, usage_type, threshold_type, threshold, is_enabled,
    name, description, priority, notification_channels
) VALUES 
(0, 'traffic', 'percentage', 80.00, false, 'Traffic 80% Alert', 'Alert when traffic usage reaches 80%', 'medium', '[]'),
(0, 'traffic', 'percentage', 90.00, false, 'Traffic 90% Alert', 'Alert when traffic usage reaches 90%', 'high', '[]'),
(0, 'traffic', 'percentage', 95.00, false, 'Traffic 95% Alert', 'Alert when traffic usage reaches 95%', 'high', '[]');

-- ==============================================================================
-- END OF UNIFIED MIGRATION
-- ==============================================================================