-- Up Migration: add_statuses_to_clients_n_sources
-- Type: tenant
-- Created: 2026-02-26 02:30:06

-- Добавляем статус для clients
DO $$ BEGIN
    CREATE TYPE client_status AS ENUM ('active', 'inactive');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE clients 
ADD COLUMN IF NOT EXISTS status client_status NOT NULL DEFAULT 'active';

CREATE INDEX IF NOT EXISTS idx_clients_status ON clients(status);

COMMENT ON COLUMN clients.status IS 'active - активен, inactive - неактивен';

-- Добавляем статус для client_sources
DO $$ BEGIN
    CREATE TYPE source_status AS ENUM ('active', 'inactive');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE client_sources 
ADD COLUMN IF NOT EXISTS status source_status NOT NULL DEFAULT 'active';

CREATE INDEX IF NOT EXISTS idx_client_sources_status ON client_sources(status);

COMMENT ON COLUMN client_sources.status IS 'active - активен, inactive - неактивен';