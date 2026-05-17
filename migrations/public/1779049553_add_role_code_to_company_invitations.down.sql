-- Down Migration: add_role_code_to_company_invitations
-- Type: public
-- Created: 2026-05-17 23:25:53

ALTER TABLE company_invitations DROP COLUMN IF EXISTS role_code;