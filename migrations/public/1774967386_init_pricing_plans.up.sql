-- Up Migration: init_pricing_plans
-- Type: public
-- Created: 2026-03-31 17:29:46

CREATE TYPE pricing_currency AS ENUM ('RUB');

CREATE TABLE pricing_plans (
    code VARCHAR(50) PRIMARY KEY,
    lvl INTEGER UNIQUE NOT NULL DEFAULT 1,
    price_per_month INTEGER NOT NULL,
    price_per_year INTEGER NOT NULL,
    price_currency pricing_currency NOT NULL DEFAULT 'RUB',
    name VARCHAR(255) NOT NULL,
    description TEXT,
    limit_db_mb INTEGER NOT NULL,
    limit_objects_mb INTEGER NOT NULL,
    limit_objects_count INTEGER NOT NULL,
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

CREATE TRIGGER update_pricing_plans_updated_at
    BEFORE UPDATE ON pricing_plans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();