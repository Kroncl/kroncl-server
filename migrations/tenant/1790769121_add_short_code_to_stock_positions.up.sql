-- Up Migration: add_short_code_to_stock_positions
-- Type: tenant
-- Created: 2026-09-30 14:52:01

-- 1. Последовательность для коротких кодов
CREATE SEQUENCE IF NOT EXISTS stock_position_short_code_seq START 1;

-- 2. Колонка
ALTER TABLE stock_positions
    ADD COLUMN short_code varchar(32);

-- 3. Заполняем существующие позиции
UPDATE stock_positions
SET short_code = 'POS-' || LPAD(nextval('stock_position_short_code_seq')::text, 6, '0')
WHERE short_code IS NULL;

-- 4. Делаем NOT NULL + уникальность
ALTER TABLE stock_positions
    ALTER COLUMN short_code SET NOT NULL;

CREATE UNIQUE INDEX idx_stock_positions_short_code
    ON stock_positions(short_code);

-- 5. Комментарий
COMMENT ON COLUMN stock_positions.short_code IS 'Короткий код позиции для печати этикеток и сканирования (например, POS-000123)';
COMMENT ON SEQUENCE stock_position_short_code_seq IS 'Последовательность коротких кодов складских позиций';