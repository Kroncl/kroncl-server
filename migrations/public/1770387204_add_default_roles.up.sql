-- Up Migration: add_default_roles
-- Type: public
-- Created: 2026-02-06 17:13:24

-- Добавляем базовые роли, если они еще не существуют
INSERT INTO roles (code, name, description, permissions)
SELECT 'admin', 'Администратор', 'Полный доступ ко всем функциям системы', '["*"]'::JSONB
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'admin');

INSERT INTO roles (code, name, description, permissions)
SELECT 'member', 'Участник', 'Стандартный участник компании с базовыми правами', '[]'::JSONB
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'member');

INSERT INTO roles (code, name, description, permissions)
SELECT 'guest', 'Гость', 'Ограниченный доступ для временных пользователей', '[]'::JSONB
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'guest');

-- Опционально: роль owner (владелец)
INSERT INTO roles (code, name, description, permissions)
SELECT 'owner', 'Владелец', 'Владелец компании с максимальными правами', '["*"]'::JSONB
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'owner');

-- Возвращаем информацию о добавленных/существующих ролях
SELECT 
    'Роли созданы/проверены:' as status,
    (SELECT COUNT(*) FROM roles WHERE code IN ('admin', 'member', 'guest', 'owner')) as total_roles,
    (SELECT string_agg(code, ', ') FROM roles WHERE code IN ('admin', 'member', 'guest', 'owner')) as role_codes;