-- Up Migration: init_company_storage
-- Created: 2026-01-21 14:19:39


-- Создаем отдельную таблицу для хранилища компании
CREATE TABLE IF NOT EXISTS company_storage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID UNIQUE NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Простое решение: только схема
    schema_name VARCHAR(50) NOT NULL,
    
    -- Статус хранилища
    status VARCHAR(20) DEFAULT 'active' 
        CHECK (status IN ('active', 'provisioning', 'failed', 'deprecated')),
    
    -- Для будущего: можно добавить тип хранилища
    storage_type VARCHAR(20) DEFAULT 'schema' 
        CHECK (storage_type IN ('schema', 'database')),
    
    -- Метаданные в JSON (для гибкости)
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Время
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Уникальный индекс на schema_name (чтобы схемы не повторялись)
    CONSTRAINT unique_schema_name UNIQUE (schema_name)
);

-- Индексы
CREATE INDEX idx_company_storage_company_id ON company_storage(company_id);
CREATE INDEX idx_company_storage_status ON company_storage(status);
CREATE INDEX idx_company_storage_created_at ON company_storage(created_at);

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_company_storage_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_company_storage_updated_at_trigger
    BEFORE UPDATE ON company_storage
    FOR EACH ROW
    EXECUTE FUNCTION update_company_storage_updated_at();

-- Функция для генерации имени схемы
CREATE OR REPLACE FUNCTION generate_tenant_schema_name(company_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    prefix VARCHAR := 'company_';
    uuid_part VARCHAR;
BEGIN
    -- Берем первые 8 символов UUID без дефисов
    uuid_part := REPLACE(SUBSTRING(company_id::text FROM 1 FOR 8), '-', '');
    RETURN prefix || uuid_part;
END;
$$ LANGUAGE plpgsql IMMUTABLE;