-- Down Migration: make_employee_fields_nullable
-- Type: tenant
-- Created: 2026-02-06

-- Удаляем новые индексы
DROP INDEX IF EXISTS idx_employees_email_unique;
DROP INDEX IF EXISTS idx_employees_email_search;

-- Восстанавливаем старый индекс
CREATE INDEX idx_employees_email ON employees(email);

-- Восстанавливаем NOT NULL ограничения
-- Сначала заполняем пустые значения
UPDATE employees 
SET 
    last_name = COALESCE(last_name, ''),
    email = COALESCE(email, ''),
    phone = COALESCE(phone, '')
WHERE 
    last_name IS NULL OR 
    email IS NULL OR 
    phone IS NULL;

-- Затем добавляем NOT NULL ограничения
ALTER TABLE employees 
    ALTER COLUMN last_name SET NOT NULL,
    ALTER COLUMN email SET NOT NULL,
    ALTER COLUMN phone SET NOT NULL;