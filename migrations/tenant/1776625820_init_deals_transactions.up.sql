-- Up Migration: init_deals_transactions
-- Type: tenant
-- Created: 2026-04-19 22:10:20

CREATE TABLE deals_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID REFERENCES deals(id) ON DELETE SET NULL,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_deals_transactions_deal_id ON deals_transactions(deal_id);
CREATE INDEX idx_deals_transactions_transaction_id ON deals_transactions(transaction_id);

CREATE OR REPLACE FUNCTION update_deals_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_deals_transactions_updated_at
    BEFORE UPDATE ON deals_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_deals_transactions_updated_at();