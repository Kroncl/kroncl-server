-- Down Migration: init_deal_positions
-- Type: tenant
-- Created: 2026-03-12 21:40:00

-- Удаляем триггер
DROP TRIGGER IF EXISTS update_deal_positions_updated_at ON deal_positions;

-- Удаляем функцию
DROP FUNCTION IF EXISTS update_deal_positions_updated_at();

-- Удаляем индексы
DROP INDEX IF EXISTS idx_deal_positions_unit_id;
DROP INDEX IF EXISTS idx_deal_positions_position_id;
DROP INDEX IF EXISTS idx_deal_positions_created_at;
DROP INDEX IF EXISTS idx_deal_positions_name;

-- Удаляем таблицу
DROP TABLE IF EXISTS deal_positions;