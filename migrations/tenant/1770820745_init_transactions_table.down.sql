-- Down Migration: init_transactions_table
-- Type: tenant
-- Created: 2026-02-11 17:39:05

DROP TABLE IF EXISTS transactions;
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_direction;
DROP TYPE IF EXISTS currency_type;