-- Up Migration: add_status_counterparties
-- Type: tenant
-- Created: 2026-02-17 03:18:51

-- Создаем enum для статуса
DO $$ BEGIN
    CREATE TYPE counterparty_status AS ENUM ('active', 'inactive');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Добавляем колонку status
ALTER TABLE counterparties 
ADD COLUMN IF NOT EXISTS status counterparty_status NOT NULL DEFAULT 'active';

-- Индекс для быстрого поиска по статусу
CREATE INDEX IF NOT EXISTS idx_counterparties_status ON counterparties(status);

-- Обновляем комментарий к таблице
COMMENT ON TABLE counterparties IS 'Контрагенты (кредиторы/дебиторы) с поддержкой статуса';