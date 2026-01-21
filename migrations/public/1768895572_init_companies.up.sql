-- Up Migration: init_companies
-- Created: 2026-01-20 10:52:52

-- Создание таблицы companies
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    avatar_url VARCHAR(255),
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индексы для оптимизации
CREATE UNIQUE INDEX idx_companies_slug ON companies(slug);
CREATE INDEX idx_companies_is_public ON companies(is_public) WHERE is_public = true;