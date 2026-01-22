-- Down Migration: add_avatar_url_to_accounts
-- Type: public
-- Created: 2026-01-22 10:45:06

-- Важно: сначала удаляем CHECK constraint, если он существует
ALTER TABLE accounts 
DROP CONSTRAINT IF EXISTS accounts_avatar_url_check;

-- Удаляем индекс, если он был создан (оставьте раскомментированным, если создавали индекс)
-- DROP INDEX IF EXISTS public.idx_accounts_avatar_url;

-- Удаляем столбец (после удаления constraint)
ALTER TABLE accounts 
DROP COLUMN IF EXISTS avatar_url;