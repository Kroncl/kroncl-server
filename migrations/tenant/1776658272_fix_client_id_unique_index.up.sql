-- Up Migration: fix_client_id_unique_index
-- Type: tenant
-- Created: 2026-04-20 07:11:12

ALTER TABLE deal_client DROP CONSTRAINT IF EXISTS deal_client_client_id_unique;