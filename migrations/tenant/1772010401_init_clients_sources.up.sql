-- Up Migration: init_clients_sources
-- Type: tenant
-- Created: 2026-02-25 12:06:41

-- Тип источника трафика
DO $$ BEGIN
    CREATE TYPE source_type AS ENUM ('organic', 'social', 'referral', 'paid', 'email', 'other');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Таблица источников трафика
CREATE TABLE client_sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    url text,
    type source_type NOT NULL DEFAULT 'other',
    comment text,
    system boolean NOT NULL DEFAULT false,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX idx_client_sources_name ON client_sources(name);
CREATE INDEX idx_client_sources_type ON client_sources(type);
CREATE INDEX idx_client_sources_system ON client_sources(system);
CREATE INDEX idx_client_sources_created_at ON client_sources(created_at DESC);

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_client_sources_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_client_sources_updated_at
    BEFORE UPDATE ON client_sources
    FOR EACH ROW
    EXECUTE FUNCTION update_client_sources_updated_at();

-- Комментарии
COMMENT ON TABLE client_sources IS 'Источники трафика для клиентов';
COMMENT ON COLUMN client_sources.type IS 'organic - органический, social - соцсети, referral - рефералы, paid - платный, email - почта, other - другое';
COMMENT ON COLUMN client_sources.system IS 'Системный источник (нельзя удалить)';