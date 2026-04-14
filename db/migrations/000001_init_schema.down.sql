-- ============================================================
-- 000001_init_schema.down.sql
-- Drops all tables in reverse dependency order.
-- Run with: migrate -path db/migrations -database $DATABASE_URL down 1
-- ============================================================

DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS cashflows;
DROP TABLE IF EXISTS point_history;
DROP TABLE IF EXISTS sale_items;
DROP TABLE IF EXISTS sales;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
