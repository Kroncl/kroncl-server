-- Down Migration: init_stock_position_batch
-- Type: tenant
-- Created: 2026-03-10 14:50:00

-- Удаляем индексы
DROP INDEX IF EXISTS idx_stock_position_batch_position_id;
DROP INDEX IF EXISTS idx_stock_position_batch_batch_id;
DROP INDEX IF EXISTS idx_stock_position_batch_created_at;

-- Удаляем таблицу
DROP TABLE IF EXISTS stock_position_batch;