-- Up Migration: drop_stock_position_batch
-- Type: tenant
-- Created: 2026-09-25

-- Связь переехала в stock_positions.income_batch_id
DROP TABLE IF EXISTS stock_position_batch;