-- Up Migration: add_default_flag_for_deal_statuses
-- Type: tenant
-- Created: 2026-03-16 14:55:47

-- 1. Добавляем колонку is_default с дефолтным значением false
ALTER TABLE deal_statuses 
ADD COLUMN is_default boolean NOT NULL DEFAULT false;

-- 2. Создаем индекс для быстрого поиска дефолтного статуса
CREATE INDEX idx_deal_statuses_is_default ON deal_statuses(is_default) WHERE is_default = true;

-- 3. Создаем функцию-триггер для проверки уникальности дефолтного статуса
CREATE OR REPLACE FUNCTION ensure_single_default_deal_status()
RETURNS TRIGGER AS $$
BEGIN
    -- Если новый/обновленный статус помечается как дефолтный
    IF NEW.is_default = true THEN
        -- Проверяем, существует ли уже другой дефолтный статус
        IF EXISTS (
            SELECT 1 FROM deal_statuses 
            WHERE is_default = true 
            AND id != NEW.id
        ) THEN
            RAISE EXCEPTION 'Only one deal status can be default. Status % cannot be set as default because another default status already exists.', NEW.id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 4. Создаем триггер на INSERT и UPDATE
CREATE TRIGGER trigger_ensure_single_default_deal_status
    BEFORE INSERT OR UPDATE OF is_default ON deal_statuses
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_default_deal_status();

-- 5. (Опционально) Устанавливаем первый статус как дефолтный, если таблица не пустая
-- Раскомментируй, если нужно автоматически назначить дефолтный статус
-- UPDATE deal_statuses 
-- SET is_default = true 
-- WHERE id = (
--     SELECT id FROM deal_statuses 
--     ORDER BY sort_order ASC, created_at ASC 
--     LIMIT 1
-- );

-- 6. Добавляем комментарий к новой колонке
COMMENT ON COLUMN deal_statuses.is_default IS 'Флаг дефолтного статуса (может быть только один true)';