-- Down Migration: seed_deals_transactions_categories
-- Type: tenant
-- Created: 2026-04-19 21:33:21

DELETE FROM transaction_categories WHERE slug IN ('deal-income', 'deal-expense');