-- Down Migration: init_deal_employees
-- Type: tenant
-- Created: 2026-03-12 21:30:00

-- Удаляем индексы
DROP INDEX IF EXISTS idx_deal_employees_deal_id;
DROP INDEX IF EXISTS idx_deal_employees_employee_id;
DROP INDEX IF EXISTS idx_deal_employees_created_at;

-- Удаляем таблицу
DROP TABLE IF EXISTS deal_employees;