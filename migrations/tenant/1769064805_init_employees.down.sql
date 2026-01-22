-- Down Migration: init_employees
-- Type: tenant
-- Created: 2026-01-22 09:53:25

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем таблицу
DROP TABLE IF EXISTS employees;