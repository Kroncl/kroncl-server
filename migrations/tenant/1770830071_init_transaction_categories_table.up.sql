-- Up Migration: init_transaction_categories_table
-- Type: tenant
-- Created: 2026-02-11 20:14:31

-- Создаем enum для направления категории (используем существующий тип или создаем новый)
DO $$ BEGIN
    CREATE TYPE category_direction AS ENUM ('income', 'expense');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS transaction_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    description text,
    direction category_direction NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_transaction_categories_name ON transaction_categories(name);
CREATE INDEX IF NOT EXISTS idx_transaction_categories_direction ON transaction_categories(direction);
CREATE INDEX IF NOT EXISTS idx_transaction_categories_created_at ON transaction_categories(created_at DESC);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_transaction_categories_updated_at ON transaction_categories;
CREATE TRIGGER update_transaction_categories_updated_at
    BEFORE UPDATE ON transaction_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();