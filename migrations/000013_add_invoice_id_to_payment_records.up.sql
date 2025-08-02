-- Add invoice_id field to payment_records table
-- This supports the core business flow: Order ’ Invoice ’ Payment ’ Service Activation

ALTER TABLE payment_records 
ADD COLUMN invoice_id BIGINT UNSIGNED NULL 
AFTER subscription_order_id;

-- Add index for efficient invoice-based payment queries
ALTER TABLE payment_records 
ADD INDEX idx_payment_records_invoice_id (invoice_id);

-- Add foreign key constraint to ensure data integrity
-- Note: This assumes invoices table exists or will be created
-- ALTER TABLE payment_records 
-- ADD CONSTRAINT fk_payment_records_invoice_id 
-- FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE SET NULL;