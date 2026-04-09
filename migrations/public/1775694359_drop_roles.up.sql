-- Up Migration: drop_roles
-- Type: public
-- Created: 2026-04-09 03:25:59

-- Проверяем существование таблицы и удаляем
DROP TABLE IF EXISTS roles CASCADE;