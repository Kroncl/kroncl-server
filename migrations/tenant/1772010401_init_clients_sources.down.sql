-- Down Migration: init_clients_sources
-- Type: tenant
-- Created: 2026-02-25 12:06:41

DROP TRIGGER IF EXISTS update_client_sources_updated_at ON client_sources;
DROP FUNCTION IF EXISTS update_client_sources_updated_at();

DROP TABLE IF EXISTS client_sources;

DROP TYPE IF EXISTS source_type;