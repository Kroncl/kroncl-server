-- Down Migration: add_status_to_stock_batches

DROP INDEX IF EXISTS idx_stock_batches_status;

ALTER TABLE stock_batches DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS stock_batch_status;