-- ============================================================================
-- REMOVE ENVIRONMENT FIELD FROM PAYMENT_CONFIGS TABLE
-- ============================================================================
-- This migration removes the unused environment field from payment_configs table

-- Remove the environment column from payment_configs table
ALTER TABLE payment_configs DROP COLUMN environment;

-- Drop the index that was created for the environment column
DROP INDEX IF EXISTS idx_payment_configs_environment ON payment_configs;