-- Up Migration: add_indexes_to_admins
-- Type: public
-- Created: 2026-05-02 07:39:52

CREATE INDEX IF NOT EXISTS idx_admins_account_id ON admins(account_id);
CREATE INDEX IF NOT EXISTS idx_admins_level ON admins(level);
CREATE UNIQUE INDEX IF NOT EXISTS idx_admins_account_id_unique ON admins(account_id);