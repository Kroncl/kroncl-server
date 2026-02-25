-- Up Migration: init_client_source_pivot_table
-- Type: tenant
-- Created: 2026-02-25 12:25:33

-- Таблица связи клиентов с источниками трафика
CREATE TABLE client_source (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id uuid NOT NULL,
    source_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- Уникальность: клиент не может быть дважды привязан к одному источнику
    CONSTRAINT uk_client_source UNIQUE (client_id, source_id),
    
    -- Внешние ключи с каскадным удалением
    CONSTRAINT fk_client_source_client 
        FOREIGN KEY (client_id) 
        REFERENCES clients(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_client_source_source 
        FOREIGN KEY (source_id) 
        REFERENCES client_sources(id) 
        ON DELETE CASCADE
);

-- Индексы для быстрых запросов
CREATE INDEX idx_client_source_client_id 
    ON client_source(client_id);

CREATE INDEX idx_client_source_source_id 
    ON client_source(source_id);

CREATE INDEX idx_client_source_created_at 
    ON client_source(created_at DESC);

-- Комментарии
COMMENT ON TABLE client_source IS 'Связь клиентов с источниками трафика';