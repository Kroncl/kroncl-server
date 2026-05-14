-- Up Migration: init_pricing_promocodes
-- Type: public
-- Created: 2026-05-14 14:37:38

CREATE TABLE IF NOT EXISTS pricing_promocodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    plan_id VARCHAR(50) NOT NULL REFERENCES pricing_plans(code) ON DELETE CASCADE,
    trial_period_days INT NOT NULL DEFAULT 0 CHECK (trial_period_days >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pricing_promocodes_code ON pricing_promocodes(code);
CREATE INDEX IF NOT EXISTS idx_pricing_promocodes_plan_id ON pricing_promocodes(plan_id);

CREATE TRIGGER update_pricing_promocodes_updated_at
    BEFORE UPDATE ON pricing_promocodes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();