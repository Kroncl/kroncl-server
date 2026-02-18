-- Up Migration: add_updated_at_for_credit_counerparty
-- Type: tenant
-- Created: 2026-02-18 05:08:16

-- Добавляем колонку updated_at
ALTER TABLE credit_counterparty 
ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_credit_counterparty_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_credit_counterparty_updated_at
    BEFORE UPDATE ON credit_counterparty
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_counterparty_updated_at();