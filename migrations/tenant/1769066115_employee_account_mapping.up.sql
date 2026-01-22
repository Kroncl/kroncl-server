-- Up Migration: employee_account_mapping
-- Type: tenant
-- Created: 2026-01-22 10:15:15

-- Таблица связи сотрудник ↔ аккаунт (One-to-One)
-- Один сотрудник может быть связан только с одним аккаунтом
-- Один аккаунт может быть связан только с одним сотрудником в этой компании
CREATE TABLE employee_account (
    -- Идентификатор связи
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Ссылка на глобального пользователя (из public.accounts)
    account_id UUID NOT NULL,
    
    -- Ссылка на сотрудника компании
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Метаданные связи
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ограничения уникальности для one-to-one
    CONSTRAINT unique_account_per_employee UNIQUE (account_id),
    CONSTRAINT unique_employee_per_account UNIQUE (employee_id)
);

-- Индексы для быстрого поиска
CREATE INDEX idx_employee_account_account_id ON employee_account(account_id);
CREATE INDEX idx_employee_account_employee_id ON employee_account(employee_id);
CREATE INDEX idx_employee_account_created_at ON employee_account(created_at);

-- Функция для обновления updated_at
CREATE OR REPLACE FUNCTION update_employee_account_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для обновления updated_at
CREATE OR REPLACE TRIGGER update_employee_account_updated_at
    BEFORE UPDATE ON employee_account
    FOR EACH ROW
    EXECUTE FUNCTION update_employee_account_updated_at();