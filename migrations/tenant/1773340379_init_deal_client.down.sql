-- Down Migration: init_deal_client
-- Type: tenant
-- Created: 2026-03-12 21:33:00

-- Удаляем индексы
DROP INDEX IF EXISTS idx_deal_client_deal_id;
DROP INDEX IF EXISTS idx_deal_client_client_id;
DROP INDEX IF EXISTS idx_deal_client_created_at;

-- Удаляем таблицу
DROP TABLE IF EXISTS deal_client;