-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_ensure_single_default_payment_method ON payment_methods;
DROP TRIGGER IF EXISTS trigger_payment_methods_updated_at ON payment_methods;
DROP FUNCTION IF EXISTS ensure_single_default_payment_method();
DROP FUNCTION IF EXISTS update_payment_methods_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_payment_methods_user_gateway_default;
DROP INDEX IF EXISTS idx_payment_methods_gateway_token;
DROP INDEX IF EXISTS idx_payment_methods_deleted_at;
DROP INDEX IF EXISTS idx_payment_methods_created_at;
DROP INDEX IF EXISTS idx_payment_methods_last_used_at;
DROP INDEX IF EXISTS idx_payment_methods_last_validated_at;
DROP INDEX IF EXISTS idx_payment_methods_status;
DROP INDEX IF EXISTS idx_payment_methods_is_active;
DROP INDEX IF EXISTS idx_payment_methods_is_default;
DROP INDEX IF EXISTS idx_payment_methods_payment_token;
DROP INDEX IF EXISTS idx_payment_methods_method;
DROP INDEX IF EXISTS idx_payment_methods_gateway;
DROP INDEX IF EXISTS idx_payment_methods_type;
DROP INDEX IF EXISTS idx_payment_methods_user_id;

-- Drop the table
DROP TABLE IF EXISTS payment_methods;