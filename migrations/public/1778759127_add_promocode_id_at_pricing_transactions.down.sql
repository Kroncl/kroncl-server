-- Down Migration: add_promocode_id_at_pricing_transactions
-- Type: public
-- Created: 2026-05-14 14:45:28

ALTER TABLE pricing_transactions 
DROP COLUMN IF EXISTS promocode_id;