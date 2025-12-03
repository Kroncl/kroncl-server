-- Down Migration  
-- Добавьте SQL для отката миграции

DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP INDEX IF EXISTS idx_accounts_email;
DROP INDEX IF EXISTS idx_accounts_created_at;
DROP TABLE IF EXISTS accounts CASCADE;