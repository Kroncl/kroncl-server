-- Down Migration: init_company_storage
-- Created: 2026-01-21 14:19:39

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS update_company_storage_updated_at_trigger ON company_storage;
DROP FUNCTION IF EXISTS update_company_storage_updated_at();

-- Удаляем вспомогательную функцию
DROP FUNCTION IF EXISTS generate_tenant_schema_name(UUID);

-- Удаляем таблицу
DROP TABLE IF EXISTS company_storage CASCADE;