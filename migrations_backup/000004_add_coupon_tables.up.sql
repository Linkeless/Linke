-- Add coupon tables and related indexes
-- This migration creates the coupons and coupon_usages tables for the coupon system

-- Create coupons table
CREATE TABLE coupons (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    
    -- Core Fields
    code VARCHAR(50) NOT NULL UNIQUE COMMENT 'Coupon code',
    name VARCHAR(100) NOT NULL COMMENT 'Coupon name',
    description TEXT COMMENT 'Coupon description',
    type VARCHAR(20) NOT NULL COMMENT 'Discount type: percentage, fixed_amount',
    value DECIMAL(10,2) NOT NULL COMMENT 'Discount value',
    
    -- Usage Limits
    max_uses INT NOT NULL DEFAULT 1 COMMENT 'Maximum usage count (0 = unlimited)',
    used_count INT NOT NULL DEFAULT 0 COMMENT 'Current usage count',
    max_uses_per_user INT NOT NULL DEFAULT 1 COMMENT 'Maximum uses per user',
    
    -- Order Requirements
    min_order_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT 'Minimum order amount',
    currency VARCHAR(3) NOT NULL DEFAULT 'USD' COMMENT 'Currency code',
    
    -- Validity Period
    valid_from TIMESTAMP NULL COMMENT 'Coupon valid from date',
    valid_until TIMESTAMP NULL COMMENT 'Coupon valid until date',
    
    -- Applicable Plans (JSON array of plan IDs)
    applicable_plans TEXT COMMENT 'JSON array of applicable plan IDs',
    
    -- Status & Visibility
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT 'Coupon status: active, inactive, expired',
    is_public BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Whether coupon is publicly visible',
    created_by BIGINT UNSIGNED NOT NULL COMMENT 'User ID who created the coupon',
    
    -- Metadata
    metadata TEXT COMMENT 'Additional metadata in JSON format',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    -- Indexes
    INDEX idx_coupons_code (code),
    INDEX idx_coupons_status (status),
    INDEX idx_coupons_type (type),
    INDEX idx_coupons_is_public (is_public),
    INDEX idx_coupons_created_by (created_by),
    INDEX idx_coupons_valid_from (valid_from),
    INDEX idx_coupons_valid_until (valid_until),
    INDEX idx_coupons_created_at (created_at),
    INDEX idx_coupons_deleted_at (deleted_at),
    
    -- Check constraints (MySQL 8.0.16+)
    CONSTRAINT chk_coupons_type CHECK (type IN ('percentage', 'fixed_amount')),
    CONSTRAINT chk_coupons_status CHECK (status IN ('active', 'inactive', 'expired')),
    CONSTRAINT chk_coupons_value CHECK (value >= 0),
    CONSTRAINT chk_coupons_max_uses CHECK (max_uses >= 0),
    CONSTRAINT chk_coupons_used_count CHECK (used_count >= 0),
    CONSTRAINT chk_coupons_max_uses_per_user CHECK (max_uses_per_user >= 1),
    CONSTRAINT chk_coupons_min_order_amount CHECK (min_order_amount >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Coupon definitions and configurations';

-- Create coupon_usages table
CREATE TABLE coupon_usages (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    
    -- Foreign Keys
    coupon_id BIGINT UNSIGNED NOT NULL COMMENT 'Reference to coupons.id',
    user_id BIGINT UNSIGNED NOT NULL COMMENT 'Reference to users.id',
    subscription_order_id BIGINT UNSIGNED NOT NULL COMMENT 'Reference to subscription_orders.id',
    
    -- Usage Details
    discount_amount DECIMAL(10,2) NOT NULL COMMENT 'Actual discount amount applied',
    order_amount DECIMAL(10,2) NOT NULL COMMENT 'Original order amount',
    currency VARCHAR(3) NOT NULL COMMENT 'Currency code',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    -- Indexes
    INDEX idx_coupon_usages_coupon_id (coupon_id),
    INDEX idx_coupon_usages_user_id (user_id),
    INDEX idx_coupon_usages_subscription_order_id (subscription_order_id),
    INDEX idx_coupon_usages_created_at (created_at),
    INDEX idx_coupon_usages_deleted_at (deleted_at),
    
    -- Composite indexes for common queries
    INDEX idx_coupon_usages_coupon_user (coupon_id, user_id),
    INDEX idx_coupon_usages_user_created (user_id, created_at),
    
    -- Unique constraint to prevent duplicate usage records
    UNIQUE KEY uk_coupon_usages_order (coupon_id, subscription_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Coupon usage history and tracking';

-- Create additional indexes for performance optimization
CREATE INDEX idx_coupons_status_public ON coupons (status, is_public) COMMENT 'Optimize public coupon queries';
CREATE INDEX idx_coupons_valid_period ON coupons (valid_from, valid_until) COMMENT 'Optimize validity period queries';
CREATE INDEX idx_coupon_usages_stats ON coupon_usages (coupon_id, created_at, discount_amount) COMMENT 'Optimize usage statistics queries';