-- Up Migration: company_accounts_drop_role_id_add_role_code
-- Type: public
-- Created: 2026-04-09 04:25:35

-- Добавляем колонку role_code
ALTER TABLE company_accounts ADD COLUMN role_code VARCHAR(50) NOT NULL DEFAULT 'guest';

-- Удаляем foreign key constraint (если есть)
ALTER TABLE company_accounts DROP CONSTRAINT IF EXISTS company_accounts_role_id_fkey;

-- Удаляем колонку role_id
ALTER TABLE company_accounts DROP COLUMN IF EXISTS role_id;

-- Создаем индекс для role_code
CREATE INDEX idx_company_accounts_role_code ON company_accounts(role_code);

-- Комментарий
COMMENT ON COLUMN company_accounts.role_code IS 'Код роли пользователя в компании (guest, owner)';