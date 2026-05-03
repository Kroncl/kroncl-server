-- Down Migration: add_indexes_to_admins
-- Type: public
-- Created: 2026-05-02 07:39:52

DROP INDEX IF EXISTS idx_admins_account_id;
DROP INDEX IF EXISTS idx_admins_level;
DROP INDEX IF EXISTS idx_admins_account_id_unique;