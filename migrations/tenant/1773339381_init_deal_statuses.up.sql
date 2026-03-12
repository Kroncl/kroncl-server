-- Up Migration: init_deal_statuses
-- Type: tenant
-- Created: 2026-03-12 21:18:00

-- Таблица статусов сделок
CREATE TABLE deal_statuses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    comment text,
    sort_order integer NOT NULL DEFAULT 1,
    color varchar(7), -- HEX-код цвета (например, #FF5733)
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    -- Проверка, что color соответствует HEX-формату
    CONSTRAINT deal_statuses_color_check CHECK (
        color IS NULL OR 
        color ~ '^#[0-9A-Fa-f]{6}$'
    )
);

-- Индексы
CREATE INDEX idx_deal_statuses_name ON deal_statuses(name);
CREATE INDEX idx_deal_statuses_sort_order ON deal_statuses(sort_order);
CREATE INDEX idx_deal_statuses_created_at ON deal_statuses(created_at DESC);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_deal_statuses_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_deal_statuses_updated_at
    BEFORE UPDATE ON deal_statuses
    FOR EACH ROW
    EXECUTE FUNCTION update_deal_statuses_updated_at();

-- Комментарии
COMMENT ON TABLE deal_statuses IS 'Статусы сделок (новый, в работе, завершен и т.д.)';
COMMENT ON COLUMN deal_statuses.name IS 'Название статуса';
COMMENT ON COLUMN deal_statuses.comment IS 'Описание/комментарий';
COMMENT ON COLUMN deal_statuses.sort_order IS 'Порядок сортировки (по умолчанию 1)';
COMMENT ON COLUMN deal_statuses.color IS 'HEX-код цвета для визуализации (например, #FF5733)';
COMMENT ON COLUMN deal_statuses.created_at IS 'Дата создания';
COMMENT ON COLUMN deal_statuses.updated_at IS 'Дата последнего обновления';