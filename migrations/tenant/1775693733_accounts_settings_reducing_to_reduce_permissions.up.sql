-- Up Migration: accounts_settings_reducing_to_reduce_permissions
-- Type: tenant
-- Created: 2026-04-09 03:15:33

-- Проверяем существование колонки reducing_permissions и переименовываем
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts_settings' AND column_name = 'reducing_permissions'
    ) THEN
        ALTER TABLE accounts_settings RENAME COLUMN reducing_permissions TO reduce_permissions;
    END IF;
END $$;

-- Обновляем комментарий
COMMENT ON COLUMN accounts_settings.reduce_permissions IS 'Список исключенных разрешений (массив строк)';