-- Up Migration: init_employees_positions
-- Type: tenant
-- Created: 2026-04-09 00:58:05

-- Таблица должностей сотрудников
CREATE TABLE employees_positions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    permissions jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Индексы
CREATE INDEX idx_employees_positions_name ON employees_positions(name);
CREATE INDEX idx_employees_positions_created_at ON employees_positions(created_at DESC);

-- GIN индекс для поиска по разрешениям
CREATE INDEX idx_employees_positions_permissions ON employees_positions USING gin(permissions);

-- Комментарии
COMMENT ON TABLE employees_positions IS 'Должности сотрудников';
COMMENT ON COLUMN employees_positions.id IS 'Уникальный идентификатор должности';
COMMENT ON COLUMN employees_positions.name IS 'Название должности';
COMMENT ON COLUMN employees_positions.description IS 'Описание должности';
COMMENT ON COLUMN employees_positions.permissions IS 'Разрешения должности (массив строк)';
COMMENT ON COLUMN employees_positions.created_at IS 'Дата создания';
COMMENT ON COLUMN employees_positions.updated_at IS 'Дата последнего обновления';