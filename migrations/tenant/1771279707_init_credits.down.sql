-- Down Migration: init_credits
-- Type: tenant
-- Created: 2026-02-17 01:08:27

DROP TRIGGER IF EXISTS update_credits_updated_at ON credits;
DROP FUNCTION IF EXISTS update_credits_updated_at();

DROP TABLE IF EXISTS credits;

DROP TYPE IF EXISTS credit_status;
DROP TYPE IF EXISTS credit_type;