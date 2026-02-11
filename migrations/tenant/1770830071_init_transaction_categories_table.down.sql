-- Down Migration: init_transaction_categories_table
-- Type: tenant
-- Created: 2026-02-11 20:14:31

DROP TABLE IF EXISTS transaction_categories;
DROP TYPE IF EXISTS category_direction;