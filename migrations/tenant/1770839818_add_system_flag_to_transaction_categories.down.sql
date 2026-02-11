-- Down Migration: add_system_flag_to_transaction_categories
-- Type: tenant
-- Created: 2026-02-11 22:56:58

ALTER TABLE transaction_categories 
DROP COLUMN IF EXISTS system;