-- Down Migration: add_default_roles (safe version)
-- Type: public
-- Created: 2026-02-06 17:13:24

-- Безопасная версия: только удаляем разрешения, но не сами роли
-- Так как роли могут использоваться в других местах

UPDATE roles 
SET permissions = '[]'::JSONB
WHERE code IN ('admin', 'member', 'guest', 'owner')
AND permissions != '[]'::JSONB;

-- Или просто логируем, что роли не удаляются
RAISE NOTICE 'Роли не удалены, так как могут использоваться в системе.';
RAISE NOTICE 'Для полного удаления ролей убедитесь, что они не используются в company_accounts.';