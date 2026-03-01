-- Down Migration: init_catalog_categories
-- Type: tenant
-- Created: 2026-02-28 08:09:26

-- Удаляем триггер
DROP TRIGGER IF EXISTS update_catalog_categories_updated_at ON catalog_categories;

-- Удаляем функцию
DROP FUNCTION IF EXISTS update_catalog_categories_updated_at();

-- Удаляем таблицу
DROP TABLE IF EXISTS catalog_categories;

-- Удаляем тип enum
DROP TYPE IF EXISTS catalog_category_status;