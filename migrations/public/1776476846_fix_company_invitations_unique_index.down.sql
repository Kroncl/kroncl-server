-- Down Migration: fix_company_invitations_unique_index
-- Type: public
-- Created: 2026-04-18 04:47:26

DROP INDEX IF EXISTS uq_company_invitations_email_company_waiting;

ALTER TABLE company_invitations ADD CONSTRAINT uq_company_invitations_email_company 
UNIQUE (email, company_id);