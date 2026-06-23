-- Up Migration: update_currency
-- Type: public
-- Created: 2026-06-21 08:59:16

ALTER TABLE currencies ADD COLUMN full_code VARCHAR(50);

UPDATE currencies SET full_code = 'bitcoin' WHERE id = 'BTC';
UPDATE currencies SET full_code = 'ethereum' WHERE id = 'ETH';
UPDATE currencies SET full_code = 'solana' WHERE id = 'SOL';
UPDATE currencies SET full_code = 'tether' WHERE id = 'USDT';
UPDATE currencies SET full_code = 'usd-coin' WHERE id = 'USDC';
UPDATE currencies SET full_code = 'the-open-network' WHERE id = 'TON';