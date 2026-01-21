-- Up Migration: init_confirmation_codes
-- Created: 2025-12-03 23:16:44

CREATE TABLE confirmation_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    code VARCHAR(10) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('email_confirmation', 'password_reset')),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Дополнительные поля для безопасности
    ip_address INET,
    user_agent TEXT
);

-- Индексы
CREATE INDEX idx_confirmation_codes_account_id ON confirmation_codes(account_id);
CREATE INDEX idx_confirmation_codes_code ON confirmation_codes(code);
CREATE INDEX idx_confirmation_codes_expires_at ON confirmation_codes(expires_at);

-- Уникальный индекс для предотвращения дубликатов НЕИСПОЛЬЗОВАННЫХ кодов одного типа
-- УБЕРИ used из индекса, иначе можно будет иметь много использованных кодов
CREATE UNIQUE INDEX idx_confirmation_codes_unique_active 
ON confirmation_codes(account_id, type) 
WHERE used = FALSE;

-- Функция для генерации кода (исправленная)
CREATE OR REPLACE FUNCTION generate_confirmation_code(
    account_uuid UUID,
    code_type VARCHAR,
    code_length INTEGER DEFAULT 6,
    expiry_minutes INTEGER DEFAULT 5
) RETURNS VARCHAR AS $$
DECLARE
    new_code VARCHAR;
BEGIN
    -- Генерируем случайный цифровой код
    new_code := LPAD(FLOOR(RANDOM() * POWER(10, code_length))::TEXT, code_length, '0');
    
    -- Удаляем старые активные коды того же типа
    -- Индекс WHERE used = FALSE гарантирует, что удаляем только неиспользованные
    DELETE FROM confirmation_codes 
    WHERE account_id = account_uuid 
      AND type = code_type
      AND used = FALSE;
    
    -- Вставляем новый код
    INSERT INTO confirmation_codes (account_id, code, type, expires_at)
    VALUES (
        account_uuid,
        new_code,
        code_type,
        NOW() + (expiry_minutes || ' minutes')::INTERVAL
    );
    
    RETURN new_code;
END;
$$ LANGUAGE plpgsql;

-- Функция для проверки кода (упрощенная)
CREATE OR REPLACE FUNCTION verify_confirmation_code(
    account_uuid UUID,
    input_code VARCHAR,
    code_type VARCHAR
) RETURNS BOOLEAN AS $$
BEGIN
    -- Проверяем и помечаем как использованный
    UPDATE confirmation_codes 
    SET used = TRUE
    WHERE account_id = account_uuid 
      AND code = input_code
      AND type = code_type
      AND used = FALSE
      AND expires_at > NOW();
    
    -- Возвращаем true если обновили хотя бы одну строку
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Функция для получения активного кода
CREATE OR REPLACE FUNCTION get_active_confirmation_code(
    account_uuid UUID,
    code_type VARCHAR
) RETURNS TABLE (
    code VARCHAR,
    expires_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT c.code, c.expires_at
    FROM confirmation_codes c
    WHERE c.account_id = account_uuid
      AND c.type = code_type
      AND c.used = FALSE
      AND c.expires_at > NOW()
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Функция для очистки устаревших кодов (опционально)
CREATE OR REPLACE FUNCTION cleanup_expired_confirmation_codes()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM confirmation_codes 
    WHERE expires_at < NOW() - INTERVAL '1 hour'
       OR (used = TRUE AND created_at < NOW() - INTERVAL '24 hours');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;