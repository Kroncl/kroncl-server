-- Up Migration: add_tracking_detail_to_catalog_units
-- Type: tenant
-- Created: 2026-03-10 13:00:59

-- Создаем enum для детализации учета
DO $$ BEGIN
    CREATE TYPE tracking_detail AS ENUM (
        'batch',   -- партионный учет (для массовых товаров)
        'serial'   -- серийный учет (для уникальных экземпляров)
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Добавляем колонку в catalog_units
ALTER TABLE catalog_units 
ADD COLUMN tracking_detail tracking_detail;

-- Обновляем существующие tracked товары: по умолчанию batch
UPDATE catalog_units 
SET tracking_detail = 'batch' 
WHERE inventory_type = 'tracked' AND tracking_detail IS NULL;

-- Добавляем комментарий
COMMENT ON COLUMN catalog_units.tracking_detail IS 
    'Детализация учета для tracked товаров: batch (партии) или serial (экземпляры). Для untracked и services = NULL';

-- Добавляем CHECK constraint
ALTER TABLE catalog_units 
ADD CONSTRAINT tracking_detail_required_for_tracked
    CHECK (
        (inventory_type = 'tracked' AND tracking_detail IS NOT NULL) OR
        (inventory_type = 'untracked' AND tracking_detail IS NULL)
    );

-- Создаем индекс
CREATE INDEX idx_catalog_units_tracking_detail 
ON catalog_units(tracking_detail) 
WHERE tracking_detail IS NOT NULL;

-- Обновляем существующий CHECK для tracked_type (FIFO/LIFO)
-- tracked_type остается только для batch-учета
ALTER TABLE catalog_units 
ADD CONSTRAINT tracked_type_required_for_batch
    CHECK (
        (tracking_detail = 'batch' AND tracked_type IS NOT NULL) OR
        (tracking_detail != 'batch' AND tracked_type IS NULL)
    );

-- Обновляем комментарий к tracked_type
COMMENT ON COLUMN catalog_units.tracked_type IS 
    'FIFO/LIFO - только для batch-учета. Для serial-учета = NULL';