-- Up Migration: fix_tracking_constraints
-- Type: tenant
-- Created: 2026-03-10 13:40:45

-- 1. Удаляем все старые constraints (IF EXISTS на всякий случай)
ALTER TABLE catalog_units DROP CONSTRAINT IF EXISTS tracked_type_required_for_tracked;
ALTER TABLE catalog_units DROP CONSTRAINT IF EXISTS tracking_detail_required_for_tracked;
ALTER TABLE catalog_units DROP CONSTRAINT IF EXISTS tracked_type_required_for_batch;

-- 2. Добавляем правильные constraints
ALTER TABLE catalog_units ADD CONSTRAINT tracking_detail_required_for_tracked
    CHECK (
        (inventory_type = 'tracked' AND tracking_detail IS NOT NULL) OR
        (inventory_type = 'untracked' AND tracking_detail IS NULL)
    );

ALTER TABLE catalog_units ADD CONSTRAINT tracked_type_required_for_batch
    CHECK (
        (tracking_detail = 'batch' AND tracked_type IS NOT NULL) OR
        (tracking_detail != 'batch' AND tracked_type IS NULL) OR
        (inventory_type = 'untracked' AND tracked_type IS NULL)
    );

-- 3. Обновляем комментарий
COMMENT ON COLUMN catalog_units.tracking_detail IS 
    'Детализация учета для tracked товаров: batch (партии) или serial (экземпляры). Для untracked и services = NULL';