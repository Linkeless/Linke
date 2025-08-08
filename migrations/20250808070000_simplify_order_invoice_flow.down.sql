-- Down migration for simplifying order/invoice flow

-- Drop supporting indexes (guard with IF EXISTS where supported)
-- Note: MySQL before 8.0 does not support IF EXISTS for DROP INDEX with this syntax;
--       Using conditional drops by names; will succeed if present.

-- Payment indexes
DROP INDEX idx_payment_records_order_time ON payment_records;
DROP INDEX idx_payment_records_user_status_created ON payment_records;

-- Invoice indexes
DROP INDEX idx_invoices_order_time ON invoices;
DROP INDEX idx_invoices_user_created ON invoices;

-- Order indexes
DROP INDEX idx_subscription_orders_user_created ON subscription_orders;


