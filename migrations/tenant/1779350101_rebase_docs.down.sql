-- Down Migration: rebase_docs
-- Type: tenant
-- Created: 2026-05-21 10:55:01

DROP TRIGGER IF EXISTS trg_docs_updated_at ON docs;
DROP FUNCTION IF EXISTS update_docs_updated_at();

DROP INDEX IF EXISTS idx_docs_module;
DROP INDEX IF EXISTS idx_docs_type;
DROP INDEX IF EXISTS idx_docs_object_path;
DROP INDEX IF EXISTS idx_docs_created_at;

ALTER TABLE docs 
DROP COLUMN IF EXISTS module,
DROP COLUMN IF EXISTS type;

ALTER TABLE docs RENAME TO transactions_reports;