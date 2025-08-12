-- ============================================================================
-- RESTORE ENVIRONMENT FIELD TO PAYMENT_CONFIGS TABLE
-- ============================================================================
-- This rollback migration restores the environment field to payment_configs table

-- Add back the environment column
ALTER TABLE payment_configs 
ADD COLUMN environment VARCHAR(20) DEFAULT 'production' COMMENT 'Environment (production, sandbox)';

-- Recreate the index for the environment column
CREATE INDEX idx_payment_configs_environment ON payment_configs (environment);