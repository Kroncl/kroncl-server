-- Down Migration: init_stock_positions
-- Type: tenant
-- Created: 2026-03-10 14:43:24

-- Удаляем таблицу
DROP TABLE IF EXISTS stock_positions;

-- Удаляем тип (только если он больше не используется)
DO $$ BEGIN
    -- Проверяем, используется ли тип где-то еще
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_class 
        WHERE reltype = (
            SELECT oid 
            FROM pg_type 
            WHERE typname = 'stock_position_type'
        )
    ) THEN
        DROP TYPE stock_position_type;
    END IF;
END $$;