-- Up Migration: init_accounts_settings
-- Type: tenant
-- Created: 2026-04-09 02:45:03

CREATE TABLE accounts_settings (
    account_id VARCHAR(255) NOT NULL PRIMARY KEY,
    increase_permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    reducing_permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounts_settings_account_id ON accounts_settings(account_id);
CREATE INDEX idx_accounts_settings_created_at ON accounts_settings(created_at DESC);
CREATE INDEX idx_accounts_settings_increase_permissions ON accounts_settings USING gin(increase_permissions);
CREATE INDEX idx_accounts_settings_reducing_permissions ON accounts_settings USING gin(reducing_permissions);

COMMENT ON TABLE accounts_settings IS 'Настройки аккаунта: дополнительные и исключенные разрешения';
COMMENT ON COLUMN accounts_settings.account_id IS 'ID аккаунта';
COMMENT ON COLUMN accounts_settings.increase_permissions IS 'Список дополнительных разрешений (массив строк)';
COMMENT ON COLUMN accounts_settings.reducing_permissions IS 'Список исключенных разрешений (массив строк)';
COMMENT ON COLUMN accounts_settings.created_at IS 'Дата создания';
COMMENT ON COLUMN accounts_settings.updated_at IS 'Дата последнего обновления';