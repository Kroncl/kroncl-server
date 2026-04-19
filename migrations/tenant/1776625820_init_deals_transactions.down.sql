-- Down Migration: init_deals_transactions
-- Type: tenant
-- Created: 2026-04-19 22:10:20

DROP TRIGGER IF EXISTS trg_deals_transactions_updated_at ON deals_transactions;
DROP FUNCTION IF EXISTS update_deals_transactions_updated_at();
DROP TABLE IF EXISTS deals_transactions;