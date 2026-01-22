-- Down Migration: employee_account_mapping
-- Type: tenant
-- Created: 2026-01-22 10:15:15

DROP TRIGGER IF EXISTS update_employee_account_updated_at ON employee_account;
DROP FUNCTION IF EXISTS update_employee_account_updated_at;
DROP TABLE IF EXISTS employee_account;