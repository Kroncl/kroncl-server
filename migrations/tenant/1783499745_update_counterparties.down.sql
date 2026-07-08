-- Down Migration: update_counterparties
-- Type: tenant
-- Created: 2026-07-08 11:35:45

DROP INDEX IF EXISTS idx_counterparties_inn;
DROP INDEX IF EXISTS idx_counterparties_default_currency;

ALTER TABLE counterparties
    DROP COLUMN IF EXISTS inn,
    DROP COLUMN IF EXISTS ogrn,
    DROP COLUMN IF EXISTS kpp,
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS default_currency;