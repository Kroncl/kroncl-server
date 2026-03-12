-- Down Migration: init_deal_statuses
-- Type: tenant
-- Created: 2026-03-12 21:18:00

-- Удаляем триггер
DROP TRIGGER IF EXISTS update_deal_statuses_updated_at ON deal_statuses;

-- Удаляем функцию (если не используется другими таблицами)
DROP FUNCTION IF EXISTS update_deal_statuses_updated_at();

-- Удаляем таблицу
DROP TABLE IF EXISTS deal_statuses;