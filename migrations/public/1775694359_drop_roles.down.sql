-- Down Migration: drop_roles
-- Type: public
-- Created: 2026-04-09 03:25:59

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