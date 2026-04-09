-- Down Migration: company_accounts_drop_role_id_add_role_code
-- Type: public
-- Created: 2026-04-09 04:25:35

-- Добавляем обратно колонку role_id
ALTER TABLE company_accounts ADD COLUMN role_id INTEGER;

-- Восстанавливаем foreign key (если таблица roles существует)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
        ALTER TABLE company_accounts ADD CONSTRAINT company_accounts_role_id_fkey 
            FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Делаем role_id NOT NULL после заполнения
ALTER TABLE company_accounts ALTER COLUMN role_id SET NOT NULL;

-- Удаляем индекс role_code
DROP INDEX IF EXISTS idx_company_accounts_role_code;

-- Удаляем колонку role_code
ALTER TABLE company_accounts DROP COLUMN IF EXISTS role_code;

-- Комментарий
COMMENT ON COLUMN company_accounts.role_id IS 'ID роли пользователя в компании';
