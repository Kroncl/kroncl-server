-- Up Migration: add_system_flag_to_transaction_categories
-- Type: tenant
-- Created: 2026-02-11 22:56:58

-- Добавляем колонку system с дефолтным значением false
ALTER TABLE transaction_categories 
ADD COLUMN IF NOT EXISTS system boolean NOT NULL DEFAULT false;

-- Индекс для быстрой фильтрации по системным/пользовательским категориям
CREATE INDEX IF NOT EXISTS idx_transaction_categories_system 
ON transaction_categories(system);

-- Комментарий для пояснения
COMMENT ON COLUMN transaction_categories.system IS 'Флаг системной категории (нельзя удалить/изменить)';