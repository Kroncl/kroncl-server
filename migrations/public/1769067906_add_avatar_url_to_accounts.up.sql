-- Up Migration: add_avatar_url_to_accounts
-- Type: public
-- Created: 2026-01-22 10:45:06

ALTER TABLE accounts 
ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(255) NULL;

-- Добавляем проверку на корректность URL (опционально)
-- Проверяет, что URL начинается с http:// или https://
ALTER TABLE accounts 
ADD CONSTRAINT accounts_avatar_url_check 
CHECK (
    avatar_url IS NULL OR 
    avatar_url ~ '^https?://[^\s/$.?#].[^\s]{1,255}$'
);