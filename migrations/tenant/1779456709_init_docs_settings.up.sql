-- Up Migration: init_docs_settings
-- Type: tenant
-- Created: 2026-05-22 16:31:49

CREATE TABLE IF NOT EXISTS docs_settings (
    legal_name VARCHAR(255),
    legal_address TEXT,
    inn VARCHAR(12),
    ogrn VARCHAR(15),
    bank_name VARCHAR(255),
    bank_bic VARCHAR(9),
    bank_account VARCHAR(20),
    director_name VARCHAR(255),
    accountant_name VARCHAR(255),
    warranty_terms TEXT,
    additional_terms TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE docs_settings IS 'Настройки компании для генерации документов (накладные, счета, договоры)';