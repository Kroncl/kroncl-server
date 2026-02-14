-- Up Migration: add_reverse_to_transactions
-- Type: tenant
-- Created: 2026-02-14 03:13:06

ALTER TABLE transactions 
ADD COLUMN IF NOT EXISTS reverse_to uuid DEFAULT NULL;

-- Индекс для быстрого поиска обратных транзакций
CREATE INDEX IF NOT EXISTS idx_transactions_reverse_to ON transactions(reverse_to);

-- Добавляем внешний ключ
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'transactions') THEN
        ALTER TABLE transactions 
        ADD CONSTRAINT fk_transactions_reverse_to 
        FOREIGN KEY (reverse_to) 
        REFERENCES transactions(id) 
        ON DELETE SET NULL;
    END IF;
END $$;

COMMENT ON COLUMN transactions.reverse_to IS 'Ссылка на исходную транзакцию, которую отменяет данная (сторно)';
