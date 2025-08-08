-- Remove added fields and indexes from subscription_orders table

-- Remove indexes
ALTER TABLE subscription_orders DROP INDEX idx_subscription_orders_invoiced_at;
ALTER TABLE subscription_orders DROP INDEX idx_subscription_orders_invoice_status;
ALTER TABLE subscription_orders DROP INDEX idx_subscription_orders_invoice_number;

-- Remove fields
ALTER TABLE subscription_orders DROP COLUMN metadata;
ALTER TABLE subscription_orders DROP COLUMN notes;
ALTER TABLE subscription_orders DROP COLUMN invoiced_at;
ALTER TABLE subscription_orders DROP COLUMN invoice_status;
ALTER TABLE subscription_orders DROP COLUMN invoice_number;