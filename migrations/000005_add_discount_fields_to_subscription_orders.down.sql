-- Remove discount fields from subscription_orders table
ALTER TABLE subscription_orders DROP COLUMN discount_value;
ALTER TABLE subscription_orders DROP COLUMN discount_type;