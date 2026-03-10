-- Up Migration: init_stock_batches
-- Type: tenant
-- Created: 2026-03-10 14:40:00

-- Создаем enum для направления движения
DO $$ BEGIN
    CREATE TYPE stock_direction AS ENUM ('income', 'outcome');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Таблица партий (документов движения)
CREATE TABLE stock_batches (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    direction stock_direction NOT NULL,  -- income - приход, outcome - расход
    comment text,                        -- комментарий к документу
    metadata jsonb DEFAULT '{}'::jsonb,  -- дополнительные данные
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX idx_stock_batches_direction ON stock_batches(direction);
CREATE INDEX idx_stock_batches_created_at ON stock_batches(created_at DESC);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_stock_batches_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_stock_batches_updated_at
    BEFORE UPDATE ON stock_batches
    FOR EACH ROW
    EXECUTE FUNCTION update_stock_batches_updated_at();

-- Комментарии
COMMENT ON TABLE stock_batches IS 'Партии движения товаров (приход/расход)';
COMMENT ON COLUMN stock_batches.direction IS 'Направление: income - приход на склад, outcome - расход со склада';
COMMENT ON COLUMN stock_batches.comment IS 'Комментарий к документу';
COMMENT ON COLUMN stock_batches.metadata IS 'Дополнительные данные (номер документа, поставщик, клиент и т.д.)';