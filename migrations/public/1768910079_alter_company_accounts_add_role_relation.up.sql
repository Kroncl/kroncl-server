-- Up Migration: alter_company_accounts_add_role_relation
-- Created: 2026-01-20 16:10:00

-- 1. Сначала добавим колонку role_id (пока nullable)
ALTER TABLE company_accounts 
ADD COLUMN role_id INTEGER REFERENCES roles(id) ON DELETE SET NULL;

-- 2. Создадим индекс для role_id
CREATE INDEX idx_company_accounts_role_id ON company_accounts(role_id);

-- 3. Перенесем данные из старого role в новый role_id
-- Предполагаем, что в roles есть записи с code = значениям из role
UPDATE company_accounts ca
SET role_id = r.id
FROM roles r
WHERE ca.role = r.code;

-- 4. Удаляем старую колонку role
ALTER TABLE company_accounts 
DROP COLUMN role;

-- 5. Теперь можно сделать role_id NOT NULL (если хотим)
ALTER TABLE company_accounts ALTER COLUMN role_id SET NOT NULL;