-- Up Migration: init_clients
-- Type: tenant
-- Created: 2026-02-25 07:31:11

-- Тип клиента
DO $$ BEGIN
    CREATE TYPE client_type AS ENUM ('individual', 'legal');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Таблица клиентов
CREATE TABLE clients (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name varchar(100) NOT NULL,
    last_name varchar(100),
    patronymic varchar(100),
    phone varchar(50),
    email varchar(255),
    type client_type NOT NULL DEFAULT 'individual',
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы для поиска (не уникальные)
CREATE INDEX idx_clients_first_name ON clients(first_name);
CREATE INDEX idx_clients_last_name ON clients(last_name);
CREATE INDEX idx_clients_patronymic ON clients(patronymic);
CREATE INDEX idx_clients_phone ON clients(phone);
CREATE INDEX idx_clients_email ON clients(email);
CREATE INDEX idx_clients_type ON clients(type);
CREATE INDEX idx_clients_created_at ON clients(created_at DESC);

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_clients_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_clients_updated_at
    BEFORE UPDATE ON clients
    FOR EACH ROW
    EXECUTE FUNCTION update_clients_updated_at();

-- Комментарии
COMMENT ON TABLE clients IS 'Клиенты (физические и юридические лица)';
COMMENT ON COLUMN clients.type IS 'individual - физлицо, legal - юрлицо';