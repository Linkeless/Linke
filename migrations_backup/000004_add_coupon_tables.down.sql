-- Rollback coupon tables and related indexes
-- This migration removes the coupons and coupon_usages tables

-- Drop additional indexes
DROP INDEX IF EXISTS idx_coupon_usages_stats ON coupon_usages;
DROP INDEX IF EXISTS idx_coupons_valid_period ON coupons;
DROP INDEX IF EXISTS idx_coupons_status_public ON coupons;

-- Drop coupon_usages table (drop first due to foreign key constraints)
DROP TABLE IF EXISTS coupon_usages;

-- Drop coupons table
DROP TABLE IF EXISTS coupons;