-- Up Migration: seed_deal_statuses
-- Type: tenant
-- Created: 2026-03-16 15:01:07

-- Ожидание (дефолтный статус)
INSERT INTO deal_statuses (id, name, comment, sort_order, color, is_default, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Ожидание',
    'Сделка ожидает начала обработки',
    10,
    '#ff006a',
    true,      -- это дефолтный статус
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Согласование
INSERT INTO deal_statuses (id, name, comment, sort_order, color, is_default, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Согласование',
    'Сделка находится на этапе согласования',
    20,
    '#ffcf2f', -- синий
    false,     -- не дефолтный
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- В работе
INSERT INTO deal_statuses (id, name, comment, sort_order, color, is_default, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'В работе',
    'Сделка активно выполняется',
    30,
    '#3451e4', -- зеленый
    false,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Успешно завершены
INSERT INTO deal_statuses (id, name, comment, sort_order, color, is_default, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Успешно завершены',
    'Сделка успешно завершена',
    40,
    '#2fe77b', -- темно-зеленый
    false,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Отклонены
INSERT INTO deal_statuses (id, name, comment, sort_order, color, is_default, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Отклонены',
    'Сделка отклонена',
    50,
    '#E74C3C', -- красный
    false,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);