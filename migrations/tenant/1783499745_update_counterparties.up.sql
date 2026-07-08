-- Up Migration: update_counterparties
-- Type: tenant
-- Created: 2026-07-08 11:35:45

-- Добавляем новые поля для реквизитов и валюты по умолчанию
ALTER TABLE counterparties
    ADD COLUMN IF NOT EXISTS inn VARCHAR(12),
    ADD COLUMN IF NOT EXISTS ogrn VARCHAR(15),
    ADD COLUMN IF NOT EXISTS kpp VARCHAR(9),
    ADD COLUMN IF NOT EXISTS address TEXT,
    ADD COLUMN IF NOT EXISTS default_currency VARCHAR(10);

-- Индексы для поиска
CREATE INDEX IF NOT EXISTS idx_counterparties_inn ON counterparties(inn);
CREATE INDEX IF NOT EXISTS idx_counterparties_default_currency ON counterparties(default_currency);

-- Комментарии
COMMENT ON COLUMN counterparties.inn IS 'ИНН контрагента';
COMMENT ON COLUMN counterparties.ogrn IS 'ОГРН/ОГРНИП контрагента';
COMMENT ON COLUMN counterparties.kpp IS 'КПП контрагента';
COMMENT ON COLUMN counterparties.address IS 'Юридический адрес контрагента';
COMMENT ON COLUMN counterparties.default_currency IS 'Валюта по умолчанию для расчётов с контрагентом (ISO-код)';