-- Up Migration: add_promocode_id_at_pricing_transactions
-- Type: public
-- Created: 2026-05-14 14:45:27

ALTER TABLE pricing_transactions 
ADD COLUMN promocode_id UUID REFERENCES pricing_promocodes(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_pricing_transactions_promocode_id ON pricing_transactions(promocode_id);