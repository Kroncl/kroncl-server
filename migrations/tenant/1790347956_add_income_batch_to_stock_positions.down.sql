-- Down Migration: add_income_batch_to_stock_positions

DROP INDEX IF EXISTS idx_stock_positions_income_batch_id;

ALTER TABLE stock_positions
    DROP CONSTRAINT IF EXISTS fk_stock_positions_income_batch;

ALTER TABLE stock_positions DROP COLUMN IF EXISTS income_batch_id;