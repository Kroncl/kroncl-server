-- Up Migration: add_deal_id_to_deal_positions
-- Type: tenant
-- Created: 2026-03-13 00:29:00

-- Добавляем колонку deal_id (сначала разрешаем NULL)
ALTER TABLE deal_positions ADD COLUMN deal_id uuid REFERENCES deals(id) ON DELETE CASCADE;

-- Теперь делаем колонку NOT NULL
ALTER TABLE deal_positions ALTER COLUMN deal_id SET NOT NULL;

-- Создаем индекс для новой колонки
CREATE INDEX idx_deal_positions_deal_id ON deal_positions(deal_id);

-- Комментарий
COMMENT ON COLUMN deal_positions.deal_id IS 'ID сделки, к которой относится позиция';