-- Up Migration: init_transactions_table
-- Type: tenant
-- Created: 2026-02-11 17:39:05

DO $$ BEGIN
    CREATE TYPE currency_type AS ENUM ('RUB', 'USD', 'EUR', 'KZT');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE transaction_direction AS ENUM ('income', 'expense');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS transactions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    base_amount bigint NOT NULL CHECK (base_amount != 0),
    currency currency_type NOT NULL DEFAULT 'RUB',
    direction transaction_direction NOT NULL,
    status transaction_status NOT NULL DEFAULT 'completed',
    comment text,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_direction ON transactions(direction);
CREATE INDEX IF NOT EXISTS idx_transactions_currency ON transactions(currency);