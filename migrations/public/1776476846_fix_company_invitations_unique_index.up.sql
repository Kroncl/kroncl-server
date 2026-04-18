-- Up Migration: fix_company_invitations_unique_index
-- Type: public
-- Created: 2026-04-18 04:47:26

ALTER TABLE company_invitations DROP CONSTRAINT IF EXISTS uq_company_invitations_email_company;

CREATE UNIQUE INDEX uq_company_invitations_email_company_waiting 
ON company_invitations(email, company_id) 
WHERE status = 'waiting';