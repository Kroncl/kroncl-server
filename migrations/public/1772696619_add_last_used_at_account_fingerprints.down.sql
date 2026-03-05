-- Down Migration: add_last_used_at_account_fingerprints
-- Type: public
-- Created: 2026-03-05 10:43:39

DROP INDEX IF EXISTS idx_account_fingerprints_last_used;

ALTER TABLE account_fingerprints 
DROP COLUMN IF EXISTS last_used_at;