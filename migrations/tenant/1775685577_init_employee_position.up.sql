-- Up Migration: init_employee_position
-- Type: tenant
-- Created: 2026-04-09 01:05:00

-- Таблица связи сотрудников с должностями
CREATE TABLE employee_position (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id uuid NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    position_id uuid NOT NULL REFERENCES employees_positions(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    -- Уникальность: один сотрудник может иметь одну должность только один раз
    CONSTRAINT employee_position_unique UNIQUE (employee_id, position_id)
);

-- Индексы
CREATE INDEX idx_employee_position_employee_id ON employee_position(employee_id);
CREATE INDEX idx_employee_position_position_id ON employee_position(position_id);
CREATE INDEX idx_employee_position_created_at ON employee_position(created_at DESC);

-- Комментарии
COMMENT ON TABLE employee_position IS 'Связь сотрудников с должностями';
COMMENT ON COLUMN employee_position.id IS 'Уникальный идентификатор записи';
COMMENT ON COLUMN employee_position.employee_id IS 'ID сотрудника';
COMMENT ON COLUMN employee_position.position_id IS 'ID должности';
COMMENT ON COLUMN employee_position.created_at IS 'Дата создания';
COMMENT ON COLUMN employee_position.updated_at IS 'Дата последнего обновления';