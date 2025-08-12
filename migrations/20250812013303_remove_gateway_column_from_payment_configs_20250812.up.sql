-- ============================================================================
-- REMOVE GATEWAY COLUMN FROM PAYMENT_CONFIGS TABLE
-- ============================================================================
-- This migration removes the old gateway column from payment_configs table
-- since it has been replaced by the method column

-- Remove the old gateway column from payment_configs table
ALTER TABLE payment_configs DROP COLUMN gateway;

-- Also remove the old config JSON column since configuration is now in structured fields
ALTER TABLE payment_configs DROP COLUMN config;