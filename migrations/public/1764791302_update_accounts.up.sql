-- Up Migration: update_accounts
-- Created: 2025-12-03 22:48:22

-- Добавляем столбец status
ALTER TABLE accounts 
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'waiting'
CHECK (status IN ('waiting', 'confirmed'));

-- Добавляем индекс
CREATE INDEX idx_accounts_status ON accounts(status);

-- Добавляем комментарий
COMMENT ON COLUMN accounts.status IS 'Статус аккаунта: waiting - ожидает подтверждения, confirmed - подтвержден';