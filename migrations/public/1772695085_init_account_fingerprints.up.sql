-- Up Migration: init_account_fingerprints
-- Type: public
-- Created: 2026-03-05 10:18:05

-- Create account_fingerprints junction table
CREATE TABLE account_fingerprints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    fingerprint_id UUID NOT NULL REFERENCES fingerprints(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Уникальность чтобы не было дубликатов
    CONSTRAINT account_fingerprints_unique UNIQUE (account_id, fingerprint_id)
);

-- Indexes for performance
CREATE INDEX idx_account_fingerprints_account ON account_fingerprints(account_id);
CREATE INDEX idx_account_fingerprints_fingerprint ON account_fingerprints(fingerprint_id);