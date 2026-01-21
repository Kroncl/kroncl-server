-- Down Migration: init_confirmation_codes
-- Удаление всех объектов, созданных в up миграции

-- Удаляем функции в обратном порядке
DROP FUNCTION IF EXISTS cleanup_expired_confirmation_codes();
DROP FUNCTION IF EXISTS get_active_confirmation_code(UUID, VARCHAR);
DROP FUNCTION IF EXISTS verify_confirmation_code(UUID, VARCHAR, VARCHAR);
DROP FUNCTION IF EXISTS generate_confirmation_code(UUID, VARCHAR, INTEGER, INTEGER);

-- Удаляем индексы
DROP INDEX IF EXISTS idx_confirmation_codes_unique_active;
DROP INDEX IF EXISTS idx_confirmation_codes_expires_at;
DROP INDEX IF EXISTS idx_confirmation_codes_code;
DROP INDEX IF EXISTS idx_confirmation_codes_account_id;

-- Удаляем таблицу
DROP TABLE IF EXISTS confirmation_codes;