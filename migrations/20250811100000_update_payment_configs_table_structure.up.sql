-- ============================================================================
-- UPDATE PAYMENT_CONFIGS TABLE STRUCTURE
-- ============================================================================
-- This migration updates the payment_configs table to support the new 
-- method-based configuration structure with url+pid+key pattern

-- Step 1: Add new columns for the restructured PaymentConfig
ALTER TABLE payment_configs 
ADD COLUMN method VARCHAR(50) NULL COMMENT 'Payment method identifier (epay, crypto, alipay, etc.)',
ADD COLUMN url VARCHAR(255) NULL COMMENT 'Payment gateway API URL',
ADD COLUMN pid VARCHAR(100) NULL COMMENT 'Partner/Merchant ID',
ADD COLUMN `key` VARCHAR(255) NULL COMMENT 'API Key or Secret',
ADD COLUMN notify_url VARCHAR(255) NULL COMMENT 'Webhook notification URL',
ADD COLUMN return_url VARCHAR(255) NULL COMMENT 'Payment return URL',
ADD COLUMN environment VARCHAR(20) DEFAULT 'production' COMMENT 'Environment (production, sandbox)';

-- Step 2: Migrate existing data from gateway+config JSON to method+structured fields
-- Extract data from existing gateway and config fields to populate new structure
UPDATE payment_configs 
SET 
    method = CASE 
        WHEN gateway = 'epay' THEN 'epay'
        WHEN gateway = 'crypto' THEN 'crypto'
        WHEN gateway = 'alipay' THEN 'alipay'
        WHEN gateway = 'wechat' THEN 'wechat'
        ELSE gateway
    END,
    url = CASE 
        WHEN config IS NOT NULL AND JSON_VALID(config) THEN 
            COALESCE(JSON_UNQUOTE(JSON_EXTRACT(config, '$.api_url')), 
                     JSON_UNQUOTE(JSON_EXTRACT(config, '$.url')))
        ELSE NULL
    END,
    pid = CASE 
        WHEN config IS NOT NULL AND JSON_VALID(config) THEN 
            COALESCE(JSON_UNQUOTE(JSON_EXTRACT(config, '$.partner_id')), 
                     JSON_UNQUOTE(JSON_EXTRACT(config, '$.merchant_id')),
                     JSON_UNQUOTE(JSON_EXTRACT(config, '$.pid')))
        ELSE NULL
    END,
    `key` = CASE 
        WHEN config IS NOT NULL AND JSON_VALID(config) THEN 
            COALESCE(JSON_UNQUOTE(JSON_EXTRACT(config, '$.api_key')), 
                     JSON_UNQUOTE(JSON_EXTRACT(config, '$.secret')),
                     JSON_UNQUOTE(JSON_EXTRACT(config, '$.key')))
        ELSE NULL
    END,
    notify_url = CASE 
        WHEN config IS NOT NULL AND JSON_VALID(config) THEN 
            JSON_UNQUOTE(JSON_EXTRACT(config, '$.notify_url'))
        ELSE NULL
    END,
    return_url = CASE 
        WHEN config IS NOT NULL AND JSON_VALID(config) THEN 
            JSON_UNQUOTE(JSON_EXTRACT(config, '$.return_url'))
        ELSE NULL
    END,
    environment = CASE 
        WHEN config IS NOT NULL AND JSON_VALID(config) THEN 
            COALESCE(JSON_UNQUOTE(JSON_EXTRACT(config, '$.environment')), 'production')
        ELSE 'production'
    END
WHERE config IS NOT NULL;

-- Step 3: Make the new method column NOT NULL and add unique constraint
-- Update any NULL method values with gateway value as fallback
UPDATE payment_configs SET method = gateway WHERE method IS NULL;

-- Modify the method column to be NOT NULL
ALTER TABLE payment_configs 
MODIFY COLUMN method VARCHAR(50) NOT NULL COMMENT 'Payment method identifier';

-- Step 4: Add indexes for the new structure
CREATE INDEX idx_payment_configs_method ON payment_configs (method);
CREATE INDEX idx_payment_configs_enabled_method ON payment_configs (is_enabled, method);
CREATE INDEX idx_payment_configs_environment ON payment_configs (environment);

-- Step 5: Remove the old gateway unique constraint and add method constraint
ALTER TABLE payment_configs DROP INDEX uk_payment_configs_gateway;
ALTER TABLE payment_configs ADD UNIQUE KEY uk_payment_configs_method (method);

-- Step 6: The config JSON column is kept for backward compatibility and additional metadata
-- but the core configuration is now in structured fields