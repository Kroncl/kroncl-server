-- Up Migration: init_credit_transactions
-- Type: tenant
-- Created: 2026-02-17 02:03:14

-- Таблица связи кредитов с транзакциями
CREATE TABLE credit_transactions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_id uuid NOT NULL,
    transaction_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- Уникальность: транзакция не может быть дважды привязана к одному кредиту
    CONSTRAINT uk_credit_transaction UNIQUE (credit_id, transaction_id),
    
    -- Внешние ключи
    CONSTRAINT fk_credit_transactions_credit 
        FOREIGN KEY (credit_id) 
        REFERENCES credits(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_credit_transactions_transaction 
        FOREIGN KEY (transaction_id) 
        REFERENCES transactions(id) 
        ON DELETE CASCADE
);

-- Индексы
CREATE INDEX idx_credit_transactions_credit_id ON credit_transactions(credit_id);
CREATE INDEX idx_credit_transactions_transaction_id ON credit_transactions(transaction_id);
CREATE INDEX idx_credit_transactions_created_at ON credit_transactions(created_at DESC);

-- Комментарии
COMMENT ON TABLE credit_transactions IS 'Связь кредитов с транзакциями (платежи по кредитам)';