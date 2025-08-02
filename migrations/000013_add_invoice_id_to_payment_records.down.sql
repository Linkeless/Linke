-- Remove invoice_id field from payment_records table
-- This rollback removes the invoice integration added for core business flow

-- Remove foreign key constraint if it exists
-- ALTER TABLE payment_records DROP FOREIGN KEY fk_payment_records_invoice_id;

-- Remove index
ALTER TABLE payment_records DROP INDEX idx_payment_records_invoice_id;

-- Remove column
ALTER TABLE payment_records DROP COLUMN invoice_id;