-- Up Migration: init_counterparties
-- Type: tenant
-- Created: 2026-02-17 01:55:08

-- Тип контрагента
DO $$ BEGIN
    CREATE TYPE counterparty_type AS ENUM ('bank', 'organization', 'person');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Таблица контрагентов
CREATE TABLE counterparties (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    comment text,
    type counterparty_type NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX idx_counterparties_type ON counterparties(type);
CREATE INDEX idx_counterparties_name ON counterparties(name);
CREATE INDEX idx_counterparties_created_at ON counterparties(created_at DESC);

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_counterparties_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_counterparties_updated_at
    BEFORE UPDATE ON counterparties
    FOR EACH ROW
    EXECUTE FUNCTION update_counterparties_updated_at();

-- Комментарии
COMMENT ON TABLE counterparties IS 'Контрагенты (кредиторы/дебиторы)';
COMMENT ON COLUMN counterparties.type IS 'bank - банк, organization - организация, person - физлицо';