-- Down Migration: update_accounts  
-- Created: 2025-12-03 22:48:22

-- Удаляем индекс
DROP INDEX IF EXISTS idx_accounts_status;

-- Удаляем столбец status
ALTER TABLE accounts 
DROP COLUMN IF EXISTS status;