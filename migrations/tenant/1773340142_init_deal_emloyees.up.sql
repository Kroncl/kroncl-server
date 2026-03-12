-- Up Migration: init_deal_employees
-- Type: tenant
-- Created: 2026-03-12 21:30:00

-- Таблица связи сделок с сотрудниками (ответственные/участники)
CREATE TABLE deal_employees (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id uuid NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    employee_id uuid NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- Защита от дублирования (сотрудник не может быть дважды привязан к одной сделке)
    CONSTRAINT deal_employees_deal_employee_unique UNIQUE (deal_id, employee_id)
);

-- Индексы
CREATE INDEX idx_deal_employees_deal_id ON deal_employees(deal_id);
CREATE INDEX idx_deal_employees_employee_id ON deal_employees(employee_id);
CREATE INDEX idx_deal_employees_created_at ON deal_employees(created_at DESC);

-- Комментарии
COMMENT ON TABLE deal_employees IS 'Сотрудники, участвующие в сделке (ответственные, участники)';
COMMENT ON COLUMN deal_employees.id IS 'Уникальный идентификатор записи';
COMMENT ON COLUMN deal_employees.deal_id IS 'ID сделки';
COMMENT ON COLUMN deal_employees.employee_id IS 'ID сотрудника (из employees)';
COMMENT ON COLUMN deal_employees.created_at IS 'Дата привязки сотрудника к сделке';