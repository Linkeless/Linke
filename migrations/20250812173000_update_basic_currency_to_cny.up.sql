-- Migration to change default currency from USD to CNY for basic tables
-- This migration updates only the core currency-related table columns and existing data

-- 1. Update subscription_plans table
ALTER TABLE subscription_plans
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'CNY';

UPDATE subscription_plans
SET currency = 'CNY'
WHERE currency = 'USD';

-- 2. Update subscription_orders table
ALTER TABLE subscription_orders
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'CNY';

UPDATE subscription_orders
SET currency = 'CNY'
WHERE currency = 'USD';

-- 3. Update user_subscriptions table
ALTER TABLE user_subscriptions
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'CNY';

UPDATE user_subscriptions
SET currency = 'CNY'
WHERE currency = 'USD';

-- 4. Update invoices table
ALTER TABLE invoices
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'CNY';

UPDATE invoices
SET currency = 'CNY'
WHERE currency = 'USD';

-- 5. Update coupons table
ALTER TABLE coupons
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'CNY';

UPDATE coupons
SET currency = 'CNY'
WHERE currency = 'USD';

-- 6. Update payment_configs table supported currencies
UPDATE payment_configs
SET supported_currencies = 'CNY'
WHERE supported_currencies IN ('USD', 'ALL', '*') OR supported_currencies LIKE '%USD%';

-- 7. Update payment_records table existing data
UPDATE payment_records
SET currency = 'CNY'
WHERE currency = 'USD';