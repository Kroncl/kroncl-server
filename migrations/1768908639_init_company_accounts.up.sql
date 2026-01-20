-- Up Migration: init_company_accounts
-- Created: 2026-01-20 14:30:39

CREATE TABLE company_accounts (
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member',
    permissions JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (company_id, account_id)
);

CREATE INDEX idx_company_accounts_account_id ON company_accounts(account_id);
CREATE INDEX idx_company_accounts_role ON company_accounts(role);
CREATE INDEX idx_company_accounts_permissions ON company_accounts USING gin(permissions);