-- Down Migration: seed_deal_statuses
-- Type: tenant
-- Created: 2026-03-12 22:27:00

-- Удаляем все созданные статусы по названиям
DELETE FROM deal_statuses 
WHERE name IN ('Ожидание', 'Согласование', 'В работе', 'Успешно завершены', 'Отклонены');