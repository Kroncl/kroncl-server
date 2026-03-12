-- Up Migration: seed_deal_statuses
-- Type: tenant
-- Created: 2026-03-12 22:27:00

-- Ожидание
INSERT INTO deal_statuses (id, name, comment, sort_order, color, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Ожидание',
    'Сделка ожидает начала обработки',
    10,
    '#FFA500', -- оранжевый
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Согласование
INSERT INTO deal_statuses (id, name, comment, sort_order, color, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Согласование',
    'Сделка находится на этапе согласования',
    20,
    '#3498DB', -- синий
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- В работе
INSERT INTO deal_statuses (id, name, comment, sort_order, color, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'В работе',
    'Сделка активно выполняется',
    30,
    '#2ECC71', -- зеленый
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Успешно завершены
INSERT INTO deal_statuses (id, name, comment, sort_order, color, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Успешно завершены',
    'Сделка успешно завершена',
    40,
    '#27AE60', -- темно-зеленый
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Отклонены
INSERT INTO deal_statuses (id, name, comment, sort_order, color, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Отклонены',
    'Сделка отклонена',
    50,
    '#E74C3C', -- красный
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);