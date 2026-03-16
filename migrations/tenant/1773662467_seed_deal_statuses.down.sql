-- Down Migration: seed_deal_statuses
-- Type: tenant
-- Created: 2026-03-16 15:01:07

-- Удаляем все сидированные статусы
DELETE FROM deal_statuses 
WHERE name IN (
    'Ожидание',
    'Согласование', 
    'В работе',
    'Успешно завершены',
    'Отклонены'
);