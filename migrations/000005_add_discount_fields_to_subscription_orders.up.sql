-- Add discount_type and discount_value columns to subscription_orders table
ALTER TABLE subscription_orders 
ADD COLUMN discount_type VARCHAR(20) NULL 
COMMENT 'Discount type (percentage, fixed)' 
AFTER setup_fee;

ALTER TABLE subscription_orders 
ADD COLUMN discount_value DECIMAL(10,2) NULL DEFAULT 0.00
COMMENT 'Discount value'
AFTER discount_type;