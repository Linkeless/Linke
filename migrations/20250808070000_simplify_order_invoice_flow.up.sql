-- Simplify order/invoice flow: add supporting indexes (no DB views, no FKs)
-- Description:
-- - Adds a read-optimized view `order_payment_invoice_summary` that aggregates
--   subscription_orders with their latest payment_record and latest invoice
-- - Adds composite indexes to speed up new query patterns
--
-- NOTE: We intentionally avoid destructive column drops in this step to keep
--       the migration safe. Column cleanup can be performed in a later migration
--       once application code is fully switched to the new flow.

-- ============================================================================
-- Indexes to support summary view resolutions (latest-by-time lookups)
-- ============================================================================

-- Payment records: look up latest record per order by (paid_at or created_at)
CREATE INDEX idx_payment_records_order_time
ON payment_records (subscription_order_id, paid_at, created_at, id);

-- Invoices: look up latest invoice per order by (issued_at or created_at)
CREATE INDEX idx_invoices_order_time
ON invoices (subscription_order_id, issued_at, created_at, id);

-- Listing optimizations for user-scoped queries
CREATE INDEX idx_subscription_orders_user_created
ON subscription_orders (user_id, created_at);

CREATE INDEX idx_invoices_user_created
ON invoices (user_id, created_at);

CREATE INDEX idx_payment_records_user_status_created
ON payment_records (user_id, status, created_at);

-- ============================================================================
-- Summary view removed per project convention: do not use DB views; perform
-- aggregation in business code using the above indexes.


