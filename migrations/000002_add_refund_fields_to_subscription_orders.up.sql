-- Add refund fields to subscription_orders table
-- These fields are needed to match the SubscriptionOrder model

ALTER TABLE subscription_orders 
ADD COLUMN setup_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT 'Setup fee for the order',
ADD COLUMN coupon_code VARCHAR(50) NULL COMMENT 'Coupon code used',
ADD COLUMN discount_type VARCHAR(20) NULL COMMENT 'Discount type: percentage or fixed',
ADD COLUMN discount_value DECIMAL(10,2) NULL COMMENT 'Discount value amount',
ADD COLUMN refund_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT 'Amount refunded',
ADD COLUMN refunded_at TIMESTAMP NULL COMMENT 'When the refund was processed',
ADD COLUMN refund_reason VARCHAR(255) NULL COMMENT 'Reason for the refund',
ADD COLUMN invoice_number VARCHAR(50) NULL COMMENT 'Invoice number',
ADD COLUMN invoice_status VARCHAR(20) NULL COMMENT 'Invoice status',
ADD COLUMN invoiced_at TIMESTAMP NULL COMMENT 'When the invoice was created',
ADD COLUMN metadata TEXT NULL COMMENT 'Additional metadata in JSON format',
ADD COLUMN notes TEXT NULL COMMENT 'Additional notes';

-- Add indexes for the new fields
CREATE INDEX idx_subscription_orders_coupon_code ON subscription_orders (coupon_code);
CREATE INDEX idx_subscription_orders_refunded_at ON subscription_orders (refunded_at);
CREATE INDEX idx_subscription_orders_invoice_number ON subscription_orders (invoice_number);
CREATE INDEX idx_subscription_orders_invoice_status ON subscription_orders (invoice_status);
CREATE INDEX idx_subscription_orders_invoiced_at ON subscription_orders (invoiced_at);