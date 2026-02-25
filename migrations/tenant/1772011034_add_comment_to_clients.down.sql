-- Down Migration: add_comment_to_clients
-- Type: tenant
-- Created: 2026-02-25 12:17:15

ALTER TABLE clients 
DROP COLUMN IF EXISTS comment;