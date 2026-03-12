-- Up Migration: init_deal_positions
-- Type: tenant
-- Created: 2026-03-12 21:40:00

-- Таблица позиций в сделке
CREATE TABLE deal_positions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    comment text,
    price numeric(15,2) NOT NULL CHECK (price >= 0),
    quantity numeric(15,3) NOT NULL CHECK (quantity > 0),
    unit varchar(20) NOT NULL DEFAULT 'pcs',
    unit_id uuid REFERENCES catalog_units(id) ON DELETE SET NULL,
    position_id uuid REFERENCES stock_positions(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    -- Логическая связь: если есть unit_id, то позиция может ссылаться на stock_positions
    -- (просто для информации, без жесткой логики в БД)
    CONSTRAINT deal_positions_unit_xor CHECK (
        (unit_id IS NOT NULL AND position_id IS NULL) OR
        (unit_id IS NULL AND position_id IS NOT NULL) OR
        (unit_id IS NULL AND position_id IS NULL)
    )
);

-- Индексы
CREATE INDEX idx_deal_positions_unit_id ON deal_positions(unit_id);
CREATE INDEX idx_deal_positions_position_id ON deal_positions(position_id);
CREATE INDEX idx_deal_positions_created_at ON deal_positions(created_at DESC);
CREATE INDEX idx_deal_positions_name ON deal_positions(name);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_deal_positions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_deal_positions_updated_at
    BEFORE UPDATE ON deal_positions
    FOR EACH ROW
    EXECUTE FUNCTION update_deal_positions_updated_at();

-- Комментарии
COMMENT ON TABLE deal_positions IS 'Позиции в сделке (товары/услуги)';
COMMENT ON COLUMN deal_positions.id IS 'Уникальный идентификатор позиции';
COMMENT ON COLUMN deal_positions.name IS 'Название позиции (копия из каталога или свободное)';
COMMENT ON COLUMN deal_positions.comment IS 'Комментарий к позиции';
COMMENT ON COLUMN deal_positions.price IS 'Цена за единицу';
COMMENT ON COLUMN deal_positions.quantity IS 'Количество';
COMMENT ON COLUMN deal_positions.unit IS 'Единица измерения';
COMMENT ON COLUMN deal_positions.unit_id IS 'Ссылка на товар в каталоге (если есть)';
COMMENT ON COLUMN deal_positions.position_id IS 'Ссылка на конкретную позицию на складе (для списания)';
COMMENT ON COLUMN deal_positions.created_at IS 'Дата создания';
COMMENT ON COLUMN deal_positions.updated_at IS 'Дата последнего обновления';