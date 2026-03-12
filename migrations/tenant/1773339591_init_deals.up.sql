-- Up Migration: init_deals
-- Type: tenant
-- Created: 2026-03-12 21:21:00

-- Таблица сделок
CREATE TABLE deals (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    comment text,
    type_id uuid REFERENCES deal_types(id) ON DELETE SET NULL, -- может быть NULL
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX idx_deals_type_id ON deals(type_id);
CREATE INDEX idx_deals_created_at ON deals(created_at DESC);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_deals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_deals_updated_at
    BEFORE UPDATE ON deals
    FOR EACH ROW
    EXECUTE FUNCTION update_deals_updated_at();

-- Комментарии
COMMENT ON TABLE deals IS 'Сделки';
COMMENT ON COLUMN deals.id IS 'Уникальный идентификатор сделки';
COMMENT ON COLUMN deals.comment IS 'Комментарий к сделке';
COMMENT ON COLUMN deals.type_id IS 'ID типа сделки (из deal_types). Может быть NULL.';
COMMENT ON COLUMN deals.created_at IS 'Дата создания';
COMMENT ON COLUMN deals.updated_at IS 'Дата последнего обновления';