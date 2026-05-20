-- Down Migration: init_transactions_reports
-- Type: tenant
-- Created: 2026-05-20 23:22:30

DROP TRIGGER IF EXISTS trg_transactions_reports_updated_at ON transactions_reports;
DROP FUNCTION IF EXISTS update_transactions_reports_updated_at();
DROP TABLE IF EXISTS transactions_reports;