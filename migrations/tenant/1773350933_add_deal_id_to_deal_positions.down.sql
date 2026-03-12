-- Down Migration: add_deal_id_to_deal_positions
-- Type: tenant
-- Created: 2026-03-13 00:29:00

-- Удаляем индекс
DROP INDEX IF EXISTS idx_deal_positions_deal_id;

-- Удаляем колонку
ALTER TABLE deal_positions DROP COLUMN IF EXISTS deal_id;