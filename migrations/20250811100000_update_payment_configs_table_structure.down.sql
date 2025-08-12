-- ============================================================================
-- ROLLBACK PAYMENT_CONFIGS TABLE STRUCTURE CHANGES
-- ============================================================================
-- This migration rolls back the payment_configs table to the original structure

-- Step 1: Remove the new unique constraint
ALTER TABLE payment_configs DROP INDEX uk_payment_configs_method;

-- Step 2: Add back the original gateway unique constraint
ALTER TABLE payment_configs ADD UNIQUE KEY uk_payment_configs_gateway (gateway);

-- Step 3: Remove the new indexes
DROP INDEX idx_payment_configs_method ON payment_configs;
DROP INDEX idx_payment_configs_enabled_method ON payment_configs;
DROP INDEX idx_payment_configs_environment ON payment_configs;

-- Step 4: Remove the new columns
ALTER TABLE payment_configs 
DROP COLUMN method,
DROP COLUMN url,
DROP COLUMN pid,
DROP COLUMN `key`,
DROP COLUMN notify_url,
DROP COLUMN return_url,
DROP COLUMN environment;