-- Down Migration: alter_company_accounts_add_role_relation

-- 1. Добавляем обратно старую колонку
ALTER TABLE company_accounts 
ADD COLUMN role VARCHAR(50) DEFAULT 'member';

-- 2. Переносим данные обратно
UPDATE company_accounts ca
SET role = r.code
FROM roles r
WHERE ca.role_id = r.id;

-- 3. Удаляем role_id
ALTER TABLE company_accounts 
DROP COLUMN role_id;

-- 4. Удаляем индекс
DROP INDEX IF EXISTS idx_company_accounts_role_id;