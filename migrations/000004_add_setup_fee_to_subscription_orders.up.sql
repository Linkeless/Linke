-- Add setup_fee column to subscription_orders table
ALTER TABLE subscription_orders 
ADD COLUMN setup_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00 
COMMENT 'Setup fee for the order' 
AFTER discount_amount;