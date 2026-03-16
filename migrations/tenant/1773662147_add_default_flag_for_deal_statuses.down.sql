-- Down Migration: add_default_flag_for_deal_statuses
-- Type: tenant
-- Created: 2026-03-16 14:55:47

-- 1. Удаляем триггер
DROP TRIGGER IF EXISTS trigger_ensure_single_default_deal_status ON deal_statuses;

-- 2. Удаляем функцию
DROP FUNCTION IF EXISTS ensure_single_default_deal_status();

-- 3. Удаляем индекс
DROP INDEX IF EXISTS idx_deal_statuses_is_default;

-- 4. Удаляем колонку
ALTER TABLE deal_statuses DROP COLUMN IF EXISTS is_default;