-- Up Migration: init_transaction_category_table
-- Type: tenant
-- Created: 2026-02-11 21:17:28

CREATE TABLE IF NOT EXISTS transaction_category (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id uuid NOT NULL,
    category_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- Композитный уникальный ключ (транзакция не может быть дважды привязана к одной категории)
    CONSTRAINT uk_transaction_category UNIQUE (transaction_id, category_id),
    
    -- Внешние ключи с каскадным удалением
    CONSTRAINT fk_transaction_category_transaction 
        FOREIGN KEY (transaction_id) 
        REFERENCES transactions(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_transaction_category_category 
        FOREIGN KEY (category_id) 
        REFERENCES transaction_categories(id) 
        ON DELETE CASCADE
);

-- Индексы для быстрых запросов
CREATE INDEX IF NOT EXISTS idx_transaction_category_transaction_id 
    ON transaction_category(transaction_id);

CREATE INDEX IF NOT EXISTS idx_transaction_category_category_id 
    ON transaction_category(category_id);

CREATE INDEX IF NOT EXISTS idx_transaction_category_created_at 
    ON transaction_category(created_at DESC);