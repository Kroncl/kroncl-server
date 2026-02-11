-- Up Migration: init_transaction_employee_table
-- Type: tenant
-- Created: 2026-02-11 19:07:36

CREATE TABLE IF NOT EXISTS transaction_employee (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id uuid NOT NULL,
    transaction_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- Композитный уникальный ключ (сотрудник не может быть дважды привязан к одной транзакции)
    CONSTRAINT uk_transaction_employee UNIQUE (employee_id, transaction_id),
    
    -- Внешние ключи
    CONSTRAINT fk_transaction_employee_employee 
        FOREIGN KEY (employee_id) 
        REFERENCES employees(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_transaction_employee_transaction 
        FOREIGN KEY (transaction_id) 
        REFERENCES transactions(id) 
        ON DELETE CASCADE
);

-- Индексы для быстрых запросов
CREATE INDEX IF NOT EXISTS idx_transaction_employee_employee_id 
    ON transaction_employee(employee_id);

CREATE INDEX IF NOT EXISTS idx_transaction_employee_transaction_id 
    ON transaction_employee(transaction_id);

CREATE INDEX IF NOT EXISTS idx_transaction_employee_created_at 
    ON transaction_employee(created_at DESC);