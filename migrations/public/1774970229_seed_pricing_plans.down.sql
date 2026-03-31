-- Down Migration: seed_pricing_plans
-- Type: public
-- Created: 2026-03-31 18:17:09

DELETE FROM pricing_plans WHERE code IN ('stoic', 'titan', 'financier');