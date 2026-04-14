-- ============================================================
-- 000002_add_indexes.down.sql
-- Drops the indexes added in 000002_add_indexes.up.sql
-- ============================================================

DROP INDEX IF EXISTS idx_sales_created_at;
DROP INDEX IF EXISTS idx_cashflows_date;
DROP INDEX IF EXISTS idx_stock_movements_variant;
DROP INDEX IF EXISTS idx_point_history_user;
DROP INDEX IF EXISTS idx_sales_customer;
DROP INDEX IF EXISTS idx_sale_items_sale;
