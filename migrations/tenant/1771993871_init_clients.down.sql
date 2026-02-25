-- Down Migration: init_clients
-- Type: tenant
-- Created: 2026-02-25 07:31:11

DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;
DROP FUNCTION IF EXISTS update_clients_updated_at();

DROP TABLE IF EXISTS clients;

DROP TYPE IF EXISTS client_type;