-- Drop payment retry tables and related structures

-- Drop views first
DROP VIEW IF EXISTS v_payment_retry_stats;
DROP VIEW IF EXISTS v_active_payment_retries;

-- Remove fields added to payment_records table
ALTER TABLE payment_records 
DROP INDEX IF EXISTS idx_payment_records_retry_count,
DROP INDEX IF EXISTS idx_payment_records_last_retry_at,
DROP INDEX IF EXISTS idx_payment_records_is_retry_enabled,
DROP COLUMN IF EXISTS retry_count,
DROP COLUMN IF EXISTS last_retry_at,
DROP COLUMN IF EXISTS is_retry_enabled,
DROP COLUMN IF EXISTS retry_failure_reason;

-- Drop payment retry history table
DROP TABLE IF EXISTS payment_retry_histories;

-- Drop payment retries table
DROP TABLE IF EXISTS payment_retries;