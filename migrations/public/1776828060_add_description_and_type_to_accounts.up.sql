-- Up Migration: add_description_and_type_to_accounts
-- Type: public
-- Created: 2026-04-22 06:21:00

ALTER TABLE accounts 
ADD COLUMN IF NOT EXISTS description VARCHAR(100),
ADD COLUMN IF NOT EXISTS type VARCHAR(20);