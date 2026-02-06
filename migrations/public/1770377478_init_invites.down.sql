-- Down Migration: init_invites
-- Type: public
-- Created: 2026-02-06 14:31:18

-- Удаление триггера
DROP TRIGGER IF EXISTS trg_company_invitations_updated_at ON company_invitations;

-- Удаление функции
DROP FUNCTION IF EXISTS update_company_invitations_updated_at();

-- Удаление таблицы (каскадно удалит все зависимости)
DROP TABLE IF EXISTS company_invitations CASCADE;