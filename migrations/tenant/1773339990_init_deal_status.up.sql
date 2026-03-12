-- Up Migration: init_deal_status
-- Type: tenant
-- Created: 2026-03-12 21:27:00

-- Таблица связи сделок со статусами (текущий статус сделки)
CREATE TABLE deal_status (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id uuid NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    status_id uuid NOT NULL REFERENCES deal_statuses(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- У одной сделки может быть только одна текущая запись
    CONSTRAINT deal_status_deal_id_unique UNIQUE (deal_id)
);

-- Индексы
CREATE INDEX idx_deal_status_deal_id ON deal_status(deal_id);
CREATE INDEX idx_deal_status_status_id ON deal_status(status_id);
CREATE INDEX idx_deal_status_created_at ON deal_status(created_at DESC);

-- Комментарии
COMMENT ON TABLE deal_status IS 'Текущий статус сделки (one-to-one)';
COMMENT ON COLUMN deal_status.id IS 'Уникальный идентификатор записи';
COMMENT ON COLUMN deal_status.deal_id IS 'ID сделки';
COMMENT ON COLUMN deal_status.status_id IS 'ID текущего статуса';
COMMENT ON COLUMN deal_status.created_at IS 'Дата установки статуса';