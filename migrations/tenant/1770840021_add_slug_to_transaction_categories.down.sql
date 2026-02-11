-- Down Migration: add_slug_to_transaction_categories
-- Type: tenant
-- Created: 2026-02-11 23:00:21

DROP INDEX IF EXISTS idx_transaction_categories_slug;
ALTER TABLE transaction_categories DROP COLUMN IF EXISTS slug;