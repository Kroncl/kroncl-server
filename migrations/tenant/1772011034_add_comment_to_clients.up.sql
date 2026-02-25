-- Up Migration: add_comment_to_clients
-- Type: tenant
-- Created: 2026-02-25 12:17:14

-- Добавляем колонку comment
ALTER TABLE clients 
ADD COLUMN IF NOT EXISTS comment text;

-- Комментарий к колонке
COMMENT ON COLUMN clients.comment IS 'Общий комментарий по клиенту';