-- Up Migration: seed_docs_settings
-- Type: tenant
-- Created: 2026-05-22 16:34:57

INSERT INTO docs_settings (legal_name, legal_address, inn, ogrn, bank_name, bank_bic, bank_account, director_name, accountant_name, warranty_terms, additional_terms, created_at, updated_at)
SELECT 
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM docs_settings);