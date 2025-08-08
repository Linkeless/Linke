-- Remove setup_fee column from subscription_orders table
ALTER TABLE subscription_orders 
DROP COLUMN setup_fee;