-- Down Migration: init_deal_status
-- Type: tenant
-- Created: 2026-03-12 21:27:00

-- Удаляем индексы
DROP INDEX IF EXISTS idx_deal_status_deal_id;
DROP INDEX IF EXISTS idx_deal_status_status_id;
DROP INDEX IF EXISTS idx_deal_status_created_at;

-- Удаляем таблицу
DROP TABLE IF EXISTS deal_status;