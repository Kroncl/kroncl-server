-- Up Migration: add_status_to_stock_batches
-- Type: tenant
-- Created: 2026-09-25

-- Enum статусов документа
DO $$ BEGIN
    CREATE TYPE stock_batch_status AS ENUM (
        'draft',      -- черновик, составление
        'labeled',    -- этикетки напечатаны (только для income)
        'confirmed',  -- проведено (принято на склад / отгружено)
        'cancelled'   -- отменено
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Добавляем статус с дефолтом
ALTER TABLE stock_batches
    ADD COLUMN status stock_batch_status NOT NULL DEFAULT 'draft';

-- Обновляем существующие: если что-то уже создано — считаем проведённым
UPDATE stock_batches SET status = 'confirmed' WHERE status = 'draft';

-- Индекс под фильтрацию
CREATE INDEX idx_stock_batches_status ON stock_batches(status);

-- Комментарий
COMMENT ON COLUMN stock_batches.status IS 'Статус документа движения: draft, labeled, confirmed, cancelled';