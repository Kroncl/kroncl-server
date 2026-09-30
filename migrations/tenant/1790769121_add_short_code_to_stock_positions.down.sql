-- Down Migration: add_short_code_to_stock_positions
-- Type: tenant
-- Created: 2026-09-30 14:52:01

DROP INDEX IF EXISTS idx_stock_positions_short_code;

ALTER TABLE stock_positions
    DROP COLUMN IF EXISTS short_code;

DROP SEQUENCE IF EXISTS stock_position_short_code_seq;