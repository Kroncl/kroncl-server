-- Up Migration: add_revoked_status_to_pricing_transactions
-- Type: public
-- Created: 2026-04-01 04:27:59

-- Добавляем новое значение в enum
ALTER TYPE pricing_transaction_status ADD VALUE IF NOT EXISTS 'revoked';