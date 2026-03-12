-- Up Migration: init_deal_types
-- Type: tenant
-- Created: 2026-03-12 21:15:00

-- Таблица типов сделок
CREATE TABLE deal_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    comment text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX idx_deal_types_name ON deal_types(name);
CREATE INDEX idx_deal_types_created_at ON deal_types(created_at DESC);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_deal_types_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_deal_types_updated_at
    BEFORE UPDATE ON deal_types
    FOR EACH ROW
    EXECUTE FUNCTION update_deal_types_updated_at();

-- Комментарии
COMMENT ON TABLE deal_types IS 'Типы сделок (продажа, покупка, аренда и т.д.)';
COMMENT ON COLUMN deal_types.name IS 'Название типа сделки';
COMMENT ON COLUMN deal_types.comment IS 'Описание/комментарий';
COMMENT ON COLUMN deal_types.created_at IS 'Дата создания';
COMMENT ON COLUMN deal_types.updated_at IS 'Дата последнего обновления';