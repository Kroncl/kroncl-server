-- Down Migration: init_stock_batches
-- Type: tenant
-- Created: 2026-03-10 14:40:00

-- Удаляем триггер
DROP TRIGGER IF EXISTS update_stock_batches_updated_at ON stock_batches;

-- Удаляем функцию
DROP FUNCTION IF EXISTS update_stock_batches_updated_at();

-- Удаляем таблицу
DROP TABLE IF EXISTS stock_batches;

-- Удаляем тип (только если он больше не используется)
DO $$ BEGIN
    -- Проверяем, используется ли тип где-то еще
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_class 
        WHERE reltype = (
            SELECT oid 
            FROM pg_type 
            WHERE typname = 'stock_direction'
        )
    ) THEN
        DROP TYPE stock_direction;
    END IF;
END $$;