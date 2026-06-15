-- Down Migration: init_api_keys
-- Type: public
-- Created: 2026-06-12 01:33:22

DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;
DROP TABLE IF EXISTS api_keys;