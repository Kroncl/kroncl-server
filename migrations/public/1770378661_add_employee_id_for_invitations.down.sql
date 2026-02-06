-- Down Migration: add_employee_id_for_invitations
-- Type: public
-- Created: 2026-02-06 14:51:01

-- Удаляем индекс
DROP INDEX IF EXISTS idx_company_invitations_employee_id;

-- Удаляем поле employee_id
ALTER TABLE company_invitations 
DROP COLUMN IF EXISTS employee_id;
