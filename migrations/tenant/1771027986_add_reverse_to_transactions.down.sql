-- Down Migration: add_reverse_to_transactions
-- Type: tenant
-- Created: 2026-02-14 03:13:07

ALTER TABLE transactions 
DROP CONSTRAINT IF EXISTS fk_transactions_reverse_to;

DROP INDEX IF EXISTS idx_transactions_reverse_to;

ALTER TABLE transactions 
DROP COLUMN IF EXISTS reverse_to;