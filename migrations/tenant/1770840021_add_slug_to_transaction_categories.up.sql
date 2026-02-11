-- Up Migration: add_slug_to_transaction_categories
-- Type: tenant
-- Created: 2026-02-11 23:00:21

-- Добавляем колонку slug (NOT NULL будет позже)
ALTER TABLE transaction_categories 
ADD COLUMN IF NOT EXISTS slug varchar(255);

-- Временно делаем nullable и заполняем дефолтными значениями
UPDATE transaction_categories 
SET slug = 'category-' || id 
WHERE slug IS NULL;

-- Теперь делаем NOT NULL и UNIQUE
ALTER TABLE transaction_categories 
ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_transaction_categories_slug 
ON transaction_categories(slug);

-- Комментарий
COMMENT ON COLUMN transaction_categories.slug IS 'URL-friendly unique identifier (generated on application level)';