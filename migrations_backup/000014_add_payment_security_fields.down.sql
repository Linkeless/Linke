-- Remove security enhancement fields from payment_records table
ALTER TABLE payment_records 
DROP INDEX idx_payment_records_notify_source,
DROP INDEX idx_payment_records_last_notify_time,
DROP COLUMN notify_source,
DROP COLUMN last_notify_time;