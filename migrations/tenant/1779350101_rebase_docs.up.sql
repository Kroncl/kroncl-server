-- Up Migration: rebase_docs
-- Type: tenant
-- Created: 2026-05-21 10:55:01

ALTER TABLE transactions_reports RENAME TO docs;

ALTER TABLE docs 
ADD COLUMN module VARCHAR(64),
ADD COLUMN type VARCHAR(64);

COMMENT ON TABLE docs IS 'Таблица для хранения информации о сгенерированных документах (отчёты, экспорты и т.д.)';
COMMENT ON COLUMN docs.id IS 'Уникальный идентификатор записи документа';
COMMENT ON COLUMN docs.object_path IS 'Путь к файлу документа в MinIO';
COMMENT ON COLUMN docs.comment IS 'Опциональный комментарий к документу';
COMMENT ON COLUMN docs.created_at IS 'Дата и время создания документа';
COMMENT ON COLUMN docs.updated_at IS 'Дата и время последнего обновления записи';
COMMENT ON COLUMN docs.module IS 'Модуль, сгенерировавший документ (fm, crm, wm, dm и т.д.)';
COMMENT ON COLUMN docs.type IS 'Тип документа (transactions, categories, counterparties, credits, full и т.д.)';

CREATE INDEX idx_docs_created_at ON docs(created_at DESC);
CREATE INDEX idx_docs_object_path ON docs(object_path);
CREATE INDEX idx_docs_module ON docs(module);
CREATE INDEX idx_docs_type ON docs(type);

CREATE OR REPLACE FUNCTION update_docs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_docs_updated_at
    BEFORE UPDATE ON docs
    FOR EACH ROW
    EXECUTE FUNCTION update_docs_updated_at();