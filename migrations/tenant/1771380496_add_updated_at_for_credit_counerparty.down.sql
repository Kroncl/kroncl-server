-- Down Migration: add_updated_at_for_credit_counerparty
-- Type: tenant
-- Created: 2026-02-18 05:08:16

DROP TRIGGER IF EXISTS update_credit_counterparty_updated_at ON credit_counterparty;
DROP FUNCTION IF EXISTS update_credit_counterparty_updated_at();

ALTER TABLE credit_counterparty 
DROP COLUMN IF EXISTS updated_at;