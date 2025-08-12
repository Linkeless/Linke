-- ============================================================================
-- RESTORE GATEWAY COLUMN TO PAYMENT_CONFIGS TABLE (ROLLBACK)
-- ============================================================================
-- This rollback migration restores the gateway column

-- Add the gateway column back
ALTER TABLE payment_configs 
ADD COLUMN gateway VARCHAR(50) NULL COMMENT 'Payment gateway identifier (rollback)';

-- Populate gateway from method column
UPDATE payment_configs SET gateway = method WHERE method IS NOT NULL;

-- Add back the unique constraint
ALTER TABLE payment_configs ADD UNIQUE KEY uk_payment_configs_gateway (gateway);