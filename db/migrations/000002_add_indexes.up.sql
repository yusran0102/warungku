-- ============================================================
-- 000002_add_indexes.up.sql
-- Adds indexes for common query patterns.
-- This is an example of how to add a new migration.
-- ============================================================

-- Speed up sales lookups by date (used in dashboard + cashflow)
CREATE INDEX IF NOT EXISTS idx_sales_created_at      ON sales(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_cashflows_date         ON cashflows(date DESC);
CREATE INDEX IF NOT EXISTS idx_stock_movements_variant ON stock_movements(variant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_point_history_user     ON point_history(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sales_customer         ON sales(customer_id);
CREATE INDEX IF NOT EXISTS idx_sale_items_sale        ON sale_items(sale_id);
