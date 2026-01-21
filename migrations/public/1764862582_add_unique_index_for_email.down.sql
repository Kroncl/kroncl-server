-- Down Migration: add_unique_index_for_email  
-- Created: 2025-12-04 18:36:23

-- Удаляем уникальный индекс
DROP INDEX IF EXISTS unique_lower_email;

-- Или, если использовали constraint:
-- ALTER TABLE accounts DROP CONSTRAINT IF EXISTS unique_email;