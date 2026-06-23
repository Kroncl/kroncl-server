-- Down Migration: restruct_transactions
-- Type: tenant
-- Created: 2026-06-23 05:14:20

-- 1. Конвертируем numeric обратно в bigint
ALTER TABLE transactions 
    ALTER COLUMN base_amount TYPE BIGINT USING ROUND(base_amount)::BIGINT;

-- 2. Создаём enum обратно
DO $$ BEGIN
    CREATE TYPE currency_type AS ENUM ('RUB', 'USD', 'EUR', 'KZT');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- 3. Конвертируем varchar обратно в enum
ALTER TABLE transactions ALTER COLUMN currency TYPE currency_type USING currency::currency_type;
ALTER TABLE credits ALTER COLUMN currency TYPE currency_type USING currency::currency_type;
ALTER TABLE catalog_units ALTER COLUMN currency TYPE currency_type USING currency::currency_type;

-- 4. Индекс
DROP INDEX IF EXISTS idx_transactions_currency;
CREATE INDEX IF NOT EXISTS idx_transactions_currency ON transactions(currency);