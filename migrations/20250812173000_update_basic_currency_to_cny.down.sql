-- Rollback migration to revert basic currency changes from CNY back to USD

-- 1. Revert subscription_plans table
UPDATE subscription_plans
SET currency = 'USD'
WHERE currency = 'CNY';

ALTER TABLE subscription_plans
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';

-- 2. Revert subscription_orders table
UPDATE subscription_orders
SET currency = 'USD'
WHERE currency = 'CNY';

ALTER TABLE subscription_orders
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';

-- 3. Revert user_subscriptions table
UPDATE user_subscriptions
SET currency = 'USD'
WHERE currency = 'CNY';

ALTER TABLE user_subscriptions
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';

-- 4. Revert invoices table
UPDATE invoices
SET currency = 'USD'
WHERE currency = 'CNY';

ALTER TABLE invoices
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';

-- 5. Revert coupons table
UPDATE coupons
SET currency = 'USD'
WHERE currency = 'CNY';

ALTER TABLE coupons
MODIFY COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';

-- 6. Revert payment_configs table supported currencies
UPDATE payment_configs
SET supported_currencies = 'USD'
WHERE supported_currencies = 'CNY';

-- 7. Revert payment_records table data
UPDATE payment_records
SET currency = 'USD'
WHERE currency = 'CNY';