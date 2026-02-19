-- Up Migration: init_logs
-- Type: tenant
-- Created: 2026-02-19 07:21:23

-- Таблица логов действий пользователей
CREATE TABLE logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    key varchar(255) NOT NULL,                     -- ключ действия (например: fm.transactions.create)
    status varchar(20) NOT NULL DEFAULT 'success', -- success, error, pending
    criticality int NOT NULL DEFAULT 1,            -- 1-10, где 10 - критично
    account_id uuid NOT NULL,                      -- ID аккаунта из публичной схемы
    request_id uuid,                                -- ID запроса для группировки
    user_agent text,                                -- браузер/клиент
    ip inet,                                        -- IP-адрес
    metadata jsonb DEFAULT '{}'::jsonb,             -- мета
    created_at timestamptz NOT NULL DEFAULT now(),
    
    -- Ограничения
    CONSTRAINT logs_status_check CHECK (status IN ('success', 'error', 'pending')),
    CONSTRAINT logs_criticality_check CHECK (criticality BETWEEN 1 AND 10)
);

-- Индексы для быстрого поиска
CREATE INDEX idx_logs_account_id ON logs(account_id);
CREATE INDEX idx_logs_created_at ON logs(created_at DESC);
CREATE INDEX idx_logs_key ON logs(key);
CREATE INDEX idx_logs_status ON logs(status);
CREATE INDEX idx_logs_criticality ON logs(criticality);
CREATE INDEX idx_logs_request_id ON logs(request_id);

-- Композитный индекс для частых фильтров
CREATE INDEX idx_logs_account_created ON logs(account_id, created_at DESC);
CREATE INDEX idx_logs_criticality_created ON logs(criticality, created_at DESC);

-- Комментарии
COMMENT ON TABLE logs IS 'Логи действий пользователей в тенанте';
COMMENT ON COLUMN logs.key IS 'Ключ действия (соответствует разрешению)';
COMMENT ON COLUMN logs.status IS 'Статус выполнения';
COMMENT ON COLUMN logs.criticality IS 'Степень критичности (1-10)';
COMMENT ON COLUMN logs.account_id IS 'ID аккаунта (без внешнего ключа)';
COMMENT ON COLUMN logs.request_id IS 'ID запроса для связки нескольких логов';
COMMENT ON COLUMN logs.metadata IS 'Детали (ошибка, изменения, метаданные)';