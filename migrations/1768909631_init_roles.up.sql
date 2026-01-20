-- Up Migration: init_roles
-- Created: 2026-01-20 16:00:00

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,    -- "owner", "admin", "manager"
    name VARCHAR(100) NOT NULL,          -- "Владелец", "Администратор"
    description TEXT,
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_roles_code ON roles(code);
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_permissions ON roles USING gin(permissions);

-- SEED данных
INSERT INTO roles (code, name, description, permissions) VALUES
('owner', 'Владелец', 'Владелец компании', '["*"]')