-- Down Migration: init_counterparties
-- Type: tenant
-- Created: 2026-02-17 01:55:08

DROP TRIGGER IF EXISTS update_counterparties_updated_at ON counterparties;
DROP FUNCTION IF EXISTS update_counterparties_updated_at();

DROP TABLE IF EXISTS counterparties;

DROP TYPE IF EXISTS counterparty_type;