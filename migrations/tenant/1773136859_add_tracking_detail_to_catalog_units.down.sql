-- Down Migration: add_tracking_detail_to_catalog_units
-- Type: tenant
-- Created: 2026-03-10 13:00:59

-- Удаляем новые констрейнты
ALTER TABLE catalog_units 
DROP CONSTRAINT IF EXISTS tracking_detail_required_for_tracked;

ALTER TABLE catalog_units 
DROP CONSTRAINT IF EXISTS tracked_type_required_for_batch;

-- Удаляем индекс
DROP INDEX IF EXISTS idx_catalog_units_tracking_detail;

-- Удаляем колонку
ALTER TABLE catalog_units 
DROP COLUMN IF EXISTS tracking_detail;

-- Удаляем тип
DROP TYPE IF EXISTS tracking_detail;