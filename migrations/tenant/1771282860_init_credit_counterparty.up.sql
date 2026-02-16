-- Up Migration: init_credit_counterparty
-- Type: tenant
-- Created: 2026-02-17 02:01:00

-- Таблица связи кредитов с контрагентами
CREATE TABLE credit_counterparty (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_id uuid NOT NULL,
    counterparty_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- Уникальность: один кредит не может быть дважды связан с одним контрагентом
    CONSTRAINT uk_credit_counterparty UNIQUE (credit_id, counterparty_id),
    
    -- Внешние ключи
    CONSTRAINT fk_credit_counterparty_credit 
        FOREIGN KEY (credit_id) 
        REFERENCES credits(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_credit_counterparty_counterparty 
        FOREIGN KEY (counterparty_id) 
        REFERENCES counterparties(id) 
        ON DELETE CASCADE
);

-- Индексы
CREATE INDEX idx_credit_counterparty_credit_id ON credit_counterparty(credit_id);
CREATE INDEX idx_credit_counterparty_counterparty_id ON credit_counterparty(counterparty_id);

-- Комментарии
COMMENT ON TABLE credit_counterparty IS 'Связь кредитов с контрагентами';