-- Up Migration: init_catalog_categories
-- Type: tenant
-- Created: 2026-02-28 08:09:26

-- Статус категории
DO $$ BEGIN
    CREATE TYPE catalog_category_status AS ENUM ('active', 'inactive');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Таблица категорий каталога
CREATE TABLE catalog_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    comment text,
    status catalog_category_status NOT NULL DEFAULT 'active',
    parent_id uuid REFERENCES catalog_categories(id) ON DELETE SET NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX idx_catalog_categories_parent_id ON catalog_categories(parent_id);
CREATE INDEX idx_catalog_categories_status ON catalog_categories(status);
CREATE INDEX idx_catalog_categories_created_at ON catalog_categories(created_at DESC);
CREATE INDEX idx_catalog_categories_name ON catalog_categories(name);

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_catalog_categories_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_catalog_categories_updated_at
    BEFORE UPDATE ON catalog_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_catalog_categories_updated_at();

-- Комментарии
COMMENT ON TABLE catalog_categories IS 'Категории каталога товаров/услуг';
COMMENT ON COLUMN catalog_categories.name IS 'Название категории';
COMMENT ON COLUMN catalog_categories.comment IS 'Описание/комментарий категории';
COMMENT ON COLUMN catalog_categories.status IS 'active - активна, inactive - неактивна';
COMMENT ON COLUMN catalog_categories.parent_id IS 'Ссылка на родительскую категорию';
COMMENT ON COLUMN catalog_categories.metadata IS 'Дополнительные данные';