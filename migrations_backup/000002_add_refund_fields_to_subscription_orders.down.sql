-- Remove refund fields from subscription_orders table
-- This is the rollback for adding refund-related fields

-- Drop indexes first
DROP INDEX IF EXISTS idx_subscription_orders_coupon_code ON subscription_orders;
DROP INDEX IF EXISTS idx_subscription_orders_refunded_at ON subscription_orders;
DROP INDEX IF EXISTS idx_subscription_orders_invoice_number ON subscription_orders;
DROP INDEX IF EXISTS idx_subscription_orders_invoice_status ON subscription_orders;
DROP INDEX IF EXISTS idx_subscription_orders_invoiced_at ON subscription_orders;

-- Drop columns
ALTER TABLE subscription_orders 
DROP COLUMN IF EXISTS setup_fee,
DROP COLUMN IF EXISTS coupon_code,
DROP COLUMN IF EXISTS discount_type,
DROP COLUMN IF EXISTS discount_value,
DROP COLUMN IF EXISTS refund_amount,
DROP COLUMN IF EXISTS refunded_at,
DROP COLUMN IF EXISTS refund_reason,
DROP COLUMN IF EXISTS invoice_number,
DROP COLUMN IF EXISTS invoice_status,
DROP COLUMN IF EXISTS invoiced_at,
DROP COLUMN IF EXISTS metadata,
DROP COLUMN IF EXISTS notes;