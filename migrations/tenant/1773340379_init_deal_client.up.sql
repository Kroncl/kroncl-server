-- Up Migration: init_deal_client
-- Type: tenant
-- Created: 2026-03-12 21:33:00

-- Таблица связи сделки с клиентом (one-to-one)
CREATE TABLE deal_client (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id uuid NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    client_id uuid NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- У одной сделки может быть только один клиент
    CONSTRAINT deal_client_deal_id_unique UNIQUE (deal_id),
    -- Защита от повторной привязки одного клиента (на всякий случай)
    CONSTRAINT deal_client_client_id_unique UNIQUE (client_id)
);

-- Индексы
CREATE INDEX idx_deal_client_deal_id ON deal_client(deal_id);
CREATE INDEX idx_deal_client_client_id ON deal_client(client_id);
CREATE INDEX idx_deal_client_created_at ON deal_client(created_at DESC);

-- Комментарии
COMMENT ON TABLE deal_client IS 'Клиент, участвующий в сделке (one-to-one)';
COMMENT ON COLUMN deal_client.id IS 'Уникальный идентификатор записи';
COMMENT ON COLUMN deal_client.deal_id IS 'ID сделки';
COMMENT ON COLUMN deal_client.client_id IS 'ID клиента (из clients)';
COMMENT ON COLUMN deal_client.created_at IS 'Дата привязки клиента к сделке';