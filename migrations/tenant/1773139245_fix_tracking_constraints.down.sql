-- Down Migration: fix_tracking_constraints
-- Type: tenant
-- Created: 2026-03-10 13:40:45

-- 1. Удаляем новые constraints
ALTER TABLE catalog_units DROP CONSTRAINT IF EXISTS tracking_detail_required_for_tracked;
ALTER TABLE catalog_units DROP CONSTRAINT IF EXISTS tracked_type_required_for_batch;

-- 2. Возвращаем старый constraint (если нужен)
ALTER TABLE catalog_units ADD CONSTRAINT tracked_type_required_for_tracked
    CHECK (
        (inventory_type = 'tracked' AND tracked_type IS NOT NULL) OR
        (inventory_type = 'untracked' AND tracked_type IS NULL)
    );

-- 3. Возвращаем старый комментарий
COMMENT ON COLUMN catalog_units.tracking_detail IS NULL;