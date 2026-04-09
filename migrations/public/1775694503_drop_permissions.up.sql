-- Up Migration: drop_permissions
-- Type: public
-- Created: 2026-04-09 03:28:23

-- Проверяем существование таблицы и удаляем
DROP TABLE IF EXISTS permissions CASCADE;