-- Down Migration: add_permission_cascade_delete_triggers
-- Created: 2026-01-21 11:00:00

DROP TRIGGER IF EXISTS validate_account_permissions ON company_accounts;
DROP TRIGGER IF EXISTS validate_role_permissions ON roles;
DROP TRIGGER IF EXISTS cascade_delete_permission ON permissions;

DROP FUNCTION IF EXISTS validate_permissions();
DROP FUNCTION IF EXISTS cascade_delete_permission();