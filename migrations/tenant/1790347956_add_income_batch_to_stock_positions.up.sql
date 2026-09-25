-- Up Migration: add_income_batch_to_stock_positions
-- Type: tenant
-- Created: 2026-09-25

-- 1. Добавляем колонку nullable
ALTER TABLE stock_positions
    ADD COLUMN income_batch_id uuid;

-- 2. Переносим данные из stock_position_batch
--    (связь позиция ↔ документ, где позиция пришла)
UPDATE stock_positions sp
SET income_batch_id = spb.batch_id
FROM stock_position_batch spb
WHERE spb.position_id = sp.id;

-- 3. Делаем NOT NULL + FK
--    Но сначала удалим позиции без привязки к документу, если такие есть
DELETE FROM stock_positions WHERE income_batch_id IS NULL;

ALTER TABLE stock_positions
    ALTER COLUMN income_batch_id SET NOT NULL;

ALTER TABLE stock_positions
    ADD CONSTRAINT fk_stock_positions_income_batch
    FOREIGN KEY (income_batch_id) REFERENCES stock_batches(id) ON DELETE RESTRICT;

CREATE INDEX idx_stock_positions_income_batch_id ON stock_positions(income_batch_id);

COMMENT ON COLUMN stock_positions.income_batch_id IS 'ID документа прихода, из которого пришла позиция';