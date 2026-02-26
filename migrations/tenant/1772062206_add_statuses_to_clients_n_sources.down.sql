-- Down Migration: add_statuses_to_clients_n_sources
-- Type: tenant
-- Created: 2026-02-26 02:30:06

ALTER TABLE clients 
DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS client_status;

ALTER TABLE client_sources 
DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS source_status;