-- Up Migration: init_accounts
-- Created: 2025-12-03 22:30:53

-- Создание таблицы accounts
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    auth_type VARCHAR(50) DEFAULT 'password',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email);
CREATE INDEX IF NOT EXISTS idx_accounts_created_at ON accounts(created_at);
CREATE INDEX IF NOT EXISTS idx_accounts_auth_type ON accounts(auth_type);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE TRIGGER update_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Валидационные ограничения (опционально)
ALTER TABLE accounts 
ADD CONSTRAINT check_email_length 
CHECK (LENGTH(email) >= 4 AND LENGTH(email) <= 254);

ALTER TABLE accounts 
ADD CONSTRAINT check_name_length 
CHECK (LENGTH(name) >= 2 AND LENGTH(name) <= 100);

ALTER TABLE accounts 
ADD CONSTRAINT check_auth_type_values 
CHECK (auth_type IN ('password', 'oauth', 'sso', 'ldap'));