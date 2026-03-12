-- Down Migration: init_deals
-- Type: tenant
-- Created: 2026-03-12 21:21:00

-- Удаляем триггер
DROP TRIGGER IF EXISTS update_deals_updated_at ON deals;

-- Удаляем функцию (если не используется другими таблицами)
DROP FUNCTION IF EXISTS update_deals_updated_at();

-- Удаляем таблицу
DROP TABLE IF EXISTS deals;