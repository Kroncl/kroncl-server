-- Down Migration: seed_dept_credit_transactions_category
-- Type: tenant
-- Created: 2026-02-18 00:36:28

DELETE FROM transaction_categories WHERE slug IN ('credit', 'dept');