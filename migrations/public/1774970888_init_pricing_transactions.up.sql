-- Up Migration: init_pricing_transactions
-- Type: public
-- Created: 2026-03-31 18:28:08

CREATE TYPE pricing_transaction_status AS ENUM ('success', 'pending', 'unsuccess');

CREATE TABLE pricing_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount INTEGER,
    currency pricing_currency NOT NULL DEFAULT 'RUB',
    status pricing_transaction_status NOT NULL DEFAULT 'pending',
    plan_code VARCHAR(50) REFERENCES pricing_plans(code),
    is_trial BOOLEAN NOT NULL DEFAULT true,
    next_plan_code VARCHAR(50) REFERENCES pricing_plans(code),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_pricing_transactions_updated_at
    BEFORE UPDATE ON pricing_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();