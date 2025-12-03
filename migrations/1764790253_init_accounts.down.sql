-- Down Migration: init_accounts  
-- Created: 2025-12-03 22:30:53

-- Удаление триггера
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;

-- Удаление функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление ограничений (если они были добавлены)
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS check_email_length;
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS check_name_length;
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS check_auth_type_values;

-- Удаление индексов
DROP INDEX IF EXISTS idx_accounts_email;
DROP INDEX IF EXISTS idx_accounts_created_at;
DROP INDEX IF EXISTS idx_accounts_auth_type;

-- Удаление таблицы
DROP TABLE IF EXISTS accounts;