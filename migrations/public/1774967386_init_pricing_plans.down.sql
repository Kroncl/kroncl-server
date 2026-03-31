-- Down Migration: init_pricing_plans
-- Type: public
-- Created: 2026-03-31 17:29:46

DROP TRIGGER IF EXISTS update_pricing_plans_updated_at ON pricing_plans;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS pricing_plans;
DROP TYPE IF EXISTS pricing_currency;