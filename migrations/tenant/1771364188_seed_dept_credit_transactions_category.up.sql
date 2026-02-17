-- Up Migration: seed_dept_credit_transactions_category
-- Type: tenant
-- Created: 2026-02-18 00:36:28

-- Вставляем категорию для доходов (кредиты)
INSERT INTO transaction_categories (id, name, description, direction, system, slug, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Доход по кредитам',
    'Поступления от выданных кредитов (нам должны)',
    'income',
    true,
    'credit',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Вставляем категорию для расходов (долги)
INSERT INTO transaction_categories (id, name, description, direction, system, slug, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Расход по займам',
    'Платежи по взятым кредитам (мы должны)',
    'expense',
    true,
    'dept',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);