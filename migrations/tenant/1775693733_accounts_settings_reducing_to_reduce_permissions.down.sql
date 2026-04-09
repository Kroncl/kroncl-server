-- Down Migration: accounts_settings_reduce_to_reducing_permissions
-- Type: tenant
-- Created: 2026-04-09 03:15:33

-- Возвращаем обратно
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts_settings' AND column_name = 'reduce_permissions'
    ) THEN
        ALTER TABLE accounts_settings RENAME COLUMN reduce_permissions TO reducing_permissions;
    END IF;
END $$;

-- Возвращаем старый комментарий
COMMENT ON COLUMN accounts_settings.reducing_permissions IS 'Список исключенных разрешений (массив строк)';