-- Add missing fields to subscription_orders table to match code entities

-- Add invoice-related fields
ALTER TABLE subscription_orders 
ADD COLUMN invoice_number VARCHAR(50) NULL 
COMMENT 'Invoice number'
AFTER discount_value;

ALTER TABLE subscription_orders 
ADD COLUMN invoice_status VARCHAR(20) NULL 
COMMENT 'Invoice status'
AFTER invoice_number;

ALTER TABLE subscription_orders 
ADD COLUMN invoiced_at TIMESTAMP NULL 
COMMENT 'Invoice generation time'
AFTER invoice_status;

-- Add additional data fields
ALTER TABLE subscription_orders 
ADD COLUMN notes TEXT NULL 
COMMENT 'Order notes'
AFTER invoiced_at;

ALTER TABLE subscription_orders 
ADD COLUMN metadata TEXT NULL 
COMMENT 'Additional metadata (JSON)'
AFTER notes;

-- Add indexes for better performance
ALTER TABLE subscription_orders 
ADD INDEX idx_subscription_orders_invoice_number (invoice_number);

ALTER TABLE subscription_orders 
ADD INDEX idx_subscription_orders_invoice_status (invoice_status);

ALTER TABLE subscription_orders 
ADD INDEX idx_subscription_orders_invoiced_at (invoiced_at);