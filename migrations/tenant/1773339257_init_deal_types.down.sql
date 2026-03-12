-- Down Migration: init_deal_types
-- Type: tenant
-- Created: 2026-03-12 21:15:00

-- Удаляем триггер
DROP TRIGGER IF EXISTS update_deal_types_updated_at ON deal_types;

-- Удаляем функцию (если не используется другими таблицами)
DROP FUNCTION IF EXISTS update_deal_types_updated_at();

-- Удаляем таблицу
DROP TABLE IF EXISTS deal_types;