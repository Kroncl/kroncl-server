-- Down Migration: add_status_counterparties
-- Type: tenant
-- Created: 2026-02-17 03:18:52

DROP INDEX IF EXISTS idx_counterparties_status;

ALTER TABLE counterparties 
DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS counterparty_status;