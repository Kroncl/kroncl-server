-- Up Migration: restruct_transactions
-- Type: tenant
-- Created: 2026-06-23 05:14:20

-- 1. Меняем base_amount
ALTER TABLE transactions 
    ALTER COLUMN base_amount TYPE NUMERIC(18,8);

-- 2. Снимаем DEFAULT с зависимых колонок
ALTER TABLE transactions ALTER COLUMN currency DROP DEFAULT;
ALTER TABLE credits ALTER COLUMN currency DROP DEFAULT;
ALTER TABLE catalog_units ALTER COLUMN currency DROP DEFAULT;

-- 3. Меняем тип колонок
ALTER TABLE transactions ALTER COLUMN currency TYPE VARCHAR(10);
ALTER TABLE credits ALTER COLUMN currency TYPE VARCHAR(10);
ALTER TABLE catalog_units ALTER COLUMN currency TYPE VARCHAR(10);

-- 4. Дропаем enum
DROP TYPE IF EXISTS currency_type;

-- 5. Индексы
DROP INDEX IF EXISTS idx_transactions_currency;
CREATE INDEX IF NOT EXISTS idx_transactions_currency ON transactions(currency);