-- Down Migration: init_pricing_transactions
-- Type: public
-- Created: 2026-03-31 18:28:09

DROP TRIGGER IF EXISTS update_pricing_transactions_updated_at ON pricing_transactions;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS pricing_transactions;
DROP TYPE IF EXISTS pricing_transaction_status;