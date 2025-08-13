-- Migration Rollback: Subscription Order Invoice Flow Refactor
-- Description: Remove order association fields and indexes

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS idx_payment_records_status_order ON payment_records;
DROP INDEX IF EXISTS idx_payment_records_user_order ON payment_records;
DROP INDEX IF EXISTS idx_invoices_user_order ON invoices;
DROP INDEX IF EXISTS idx_payment_records_invoice_id ON payment_records;
DROP INDEX IF EXISTS idx_payment_records_subscription_order_id ON payment_records;
DROP INDEX IF EXISTS idx_invoices_subscription_order_id ON invoices;

-- Remove columns (in reverse order of addition)
ALTER TABLE payment_records DROP COLUMN IF EXISTS invoice_id;
ALTER TABLE payment_records DROP COLUMN IF EXISTS subscription_order_id;
ALTER TABLE invoices DROP COLUMN IF EXISTS subscription_order_id;

-- Reset table comments
ALTER TABLE invoices COMMENT = '';
ALTER TABLE payment_records COMMENT = '';