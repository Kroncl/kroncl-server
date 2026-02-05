-- Up Migration: make_employee_fields_nullable
-- Type: tenant
-- Created: 2026-02-06

-- Сначала удаляем существующие ограничения если есть
ALTER TABLE employees 
    ALTER COLUMN last_name DROP NOT NULL,
    ALTER COLUMN email DROP NOT NULL,
    ALTER COLUMN phone DROP NOT NULL;

-- Опционально: обновляем существующие пустые строки в NULL
UPDATE employees 
SET 
    last_name = NULL 
WHERE last_name = '';

UPDATE employees 
SET 
    email = NULL 
WHERE email = '';

UPDATE employees 
SET 
    phone = NULL 
WHERE phone = '';

-- Индекс для email (уникальность только для NOT NULL значений)
-- Удаляем старый индекс если был уникальным
DROP INDEX IF EXISTS idx_employees_email;

-- Создаем новый индекс с условием для уникальности не-NULL email
CREATE UNIQUE INDEX idx_employees_email_unique 
ON employees(email) 
WHERE email IS NOT NULL;

-- Опциональный индекс для поиска по email
CREATE INDEX idx_employees_email_search 
ON employees(email);