-- Down Migration: add_description_and_type_to_accounts
-- Type: public
-- Created: 2026-04-22 06:21:00

ALTER TABLE accounts 
DROP COLUMN IF EXISTS description,
DROP COLUMN IF EXISTS type;