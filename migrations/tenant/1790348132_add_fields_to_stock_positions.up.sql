-- Up Migration: add_fields_to_stock_positions
-- Type: tenant
-- Created: 2026-09-25

-- unit_price (цена за единицу)
ALTER TABLE stock_positions
    ADD COLUMN unit_price numeric(15,2) NOT NULL DEFAULT 0;

-- maker (производитель)
ALTER TABLE stock_positions
    ADD COLUMN maker varchar(255);

-- barcode (ссылка на словарь)
ALTER TABLE stock_positions
    ADD COLUMN barcode_id uuid REFERENCES barcodes(id) ON DELETE SET NULL;

-- updated_at
ALTER TABLE stock_positions
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_stock_positions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_stock_positions_updated_at
    BEFORE UPDATE ON stock_positions
    FOR EACH ROW
    EXECUTE FUNCTION update_stock_positions_updated_at();

-- Индексы
CREATE INDEX idx_stock_positions_barcode_id ON stock_positions(barcode_id);

-- Комментарии
COMMENT ON COLUMN stock_positions.unit_price IS 'Цена за единицу товара';
COMMENT ON COLUMN stock_positions.maker IS 'Производитель (опционально)';
COMMENT ON COLUMN stock_positions.barcode_id IS 'Ссылка на словарь баркодов';
COMMENT ON COLUMN stock_positions.updated_at IS 'Дата последнего обновления';