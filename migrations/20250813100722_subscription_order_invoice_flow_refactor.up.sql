-- Migration: Subscription Order Invoice Flow Refactor
-- Created: 2025-01-13 10:07:22
-- Description: Add order association fields to invoices and payment_records tables
-- Following MySQL best practices: no foreign keys, no views, application-layer integrity

-- Add subscription_order_id to invoices table
ALTER TABLE invoices 
ADD COLUMN subscription_order_id BIGINT UNSIGNED NULL 
COMMENT 'Links invoice to subscription order (managed by application layer)'
AFTER user_id;

-- Add subscription_order_id and invoice_id to payment_records table  
ALTER TABLE payment_records 
ADD COLUMN subscription_order_id BIGINT UNSIGNED NULL
COMMENT 'Links payment to subscription order (managed by application layer)'
AFTER user_id;

ALTER TABLE payment_records
ADD COLUMN invoice_id BIGINT UNSIGNED NULL
COMMENT 'Links payment to invoice (managed by application layer)'
AFTER subscription_order_id;

-- Create indexes for optimal query performance
-- These indexes support the most common query patterns in the application

-- Index for finding invoices by subscription order
CREATE INDEX idx_invoices_subscription_order_id 
ON invoices(subscription_order_id);

-- Index for finding payments by subscription order
CREATE INDEX idx_payment_records_subscription_order_id 
ON payment_records(subscription_order_id);

-- Index for finding payments by invoice
CREATE INDEX idx_payment_records_invoice_id 
ON payment_records(invoice_id);

-- Composite indexes for common query patterns
-- Index for user's order-related invoices
CREATE INDEX idx_invoices_user_order 
ON invoices(user_id, subscription_order_id);

-- Index for user's order-related payments
CREATE INDEX idx_payment_records_user_order 
ON payment_records(user_id, subscription_order_id);

-- Index for order completion flow queries (status + order_id)
CREATE INDEX idx_payment_records_status_order 
ON payment_records(status, subscription_order_id);

-- Data migration: Link existing records where possible
-- This is safe to run as it only updates NULL values

-- Link existing payment records to invoices based on matching criteria
-- (amount, user_id, created within reasonable timeframe)
UPDATE payment_records pr
JOIN invoices i ON (
    i.user_id = pr.user_id 
    AND ABS(i.amount - pr.amount) < 0.01
    AND i.created_at BETWEEN DATE_SUB(pr.created_at, INTERVAL 1 HOUR) 
                         AND DATE_ADD(pr.created_at, INTERVAL 1 HOUR)
    AND pr.invoice_id IS NULL
    AND i.subscription_order_id IS NULL
)
SET pr.invoice_id = i.id
WHERE pr.subscription_order_id IS NULL;

-- Add helpful comments to the tables for documentation
ALTER TABLE invoices 
COMMENT = 'Invoice records with optional subscription order linking';

ALTER TABLE payment_records 
COMMENT = 'Payment records with subscription order and invoice linking';

-- Note: Application layer is responsible for:
-- 1. Referential integrity between orders, invoices, and payments
-- 2. Cascade deletes and updates
-- 3. Data validation and consistency checks
-- 4. Transaction management across related records