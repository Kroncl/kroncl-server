-- Up Migration: init_stock_positions
-- Type: tenant
-- Created: 2026-03-10 14:43:24

-- Создаем enum для типа позиции
DO $$ BEGIN
    CREATE TYPE stock_position_type AS ENUM ('batch', 'serial');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Таблица позиций на складе
CREATE TABLE stock_positions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type stock_position_type NOT NULL,           -- batch - партия, serial - конкретный экземпляр
    unit_id uuid NOT NULL REFERENCES catalog_units(id) ON DELETE RESTRICT,
    quantity numeric(15,3) NOT NULL CHECK (
        (type = 'batch' AND quantity > 0) OR
        (type = 'serial' AND quantity = 1)
    ),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX idx_stock_positions_type ON stock_positions(type);
CREATE INDEX idx_stock_positions_unit_id ON stock_positions(unit_id);
CREATE INDEX idx_stock_positions_created_at ON stock_positions(created_at DESC);

-- Комментарии
COMMENT ON TABLE stock_positions IS 'Текущие позиции на складе (batch/партии и serial/экземпляры)';
COMMENT ON COLUMN stock_positions.type IS 'Тип позиции: batch - партионный учет (количество > 1), serial - поштучный учет (количество всегда 1)';
COMMENT ON COLUMN stock_positions.unit_id IS 'Товарная позиция из каталога';
COMMENT ON COLUMN stock_positions.quantity IS 'Количество: для batch может быть > 1, для serial всегда = 1';
COMMENT ON COLUMN stock_positions.created_at IS 'Дата поступления на склад';