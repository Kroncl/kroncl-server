-- Down Migration: drop_permissions
-- Type: public
-- Created: 2026-04-09 03:28:23

-- Таблица зарегистрированных разрешений в системе
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,  -- "crm.clients.create"
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индекс для быстрого поиска по коду
CREATE INDEX idx_permissions_code ON permissions(code);

-- SEED: Базовые разрешения системы
INSERT INTO permissions (code, description) VALUES

-- Полный доступ
('*', 'Полный доступ ко всему')
