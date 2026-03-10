-- Up Migration: init_stock_position_batch
-- Type: tenant
-- Created: 2026-03-10 14:50:00

-- Таблица связи позиций с партиями (документами движения)
CREATE TABLE stock_position_batch (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    position_id uuid NOT NULL REFERENCES stock_positions(id) ON DELETE CASCADE,
    batch_id uuid NOT NULL REFERENCES stock_batches(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- Защита от дублирования связей
    UNIQUE(position_id, batch_id)
);

-- Индексы для быстрого поиска
CREATE INDEX idx_stock_position_batch_position_id ON stock_position_batch(position_id);
CREATE INDEX idx_stock_position_batch_batch_id ON stock_position_batch(batch_id);
CREATE INDEX idx_stock_position_batch_created_at ON stock_position_batch(created_at DESC);

-- Комментарии
COMMENT ON TABLE stock_position_batch IS 'Связь между позициями на складе и партиями движения';
COMMENT ON COLUMN stock_position_batch.position_id IS 'ID позиции на складе (из stock_positions)';
COMMENT ON COLUMN stock_position_batch.batch_id IS 'ID партии движения (из stock_batches)';
COMMENT ON COLUMN stock_position_batch.created_at IS 'Дата создания связи';