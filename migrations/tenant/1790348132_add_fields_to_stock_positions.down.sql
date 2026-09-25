-- Down Migration: add_fields_to_stock_positions

DROP TRIGGER IF EXISTS update_stock_positions_updated_at ON stock_positions;
DROP FUNCTION IF EXISTS update_stock_positions_updated_at();

DROP INDEX IF EXISTS idx_stock_positions_barcode_id;

ALTER TABLE stock_positions
    DROP COLUMN IF EXISTS unit_price,
    DROP COLUMN IF EXISTS maker,
    DROP COLUMN IF EXISTS barcode_id,
    DROP COLUMN IF EXISTS updated_at;