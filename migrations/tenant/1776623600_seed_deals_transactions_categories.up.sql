-- Up Migration: seed_deals_transactions_categories
-- Type: tenant
-- Created: 2026-04-19 21:33:20

-- Вставляем категорию для доходов по сделкам
INSERT INTO transaction_categories (id, name, description, direction, system, slug, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Доход со сделок',
    'Поступления прибыли со сделок',
    'income',
    true,
    'deal-income',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Вставляем категорию для расходов по сделкам
INSERT INTO transaction_categories (id, name, description, direction, system, slug, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Расход на сделки',
    'Операционные расходы на сделки',
    'expense',
    true,
    'deal-expense',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);