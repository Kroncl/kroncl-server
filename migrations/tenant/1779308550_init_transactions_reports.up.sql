-- Up Migration: init_transactions_reports
-- Type: tenant
-- Created: 2026-05-20 23:22:30

-- Up Migration: init_transactions_reports
-- Type: tenant
-- Created: 2026-05-20 23:22:30

CREATE TABLE IF NOT EXISTS transactions_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_path VARCHAR(512) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_reports_created_at ON transactions_reports(created_at DESC);
CREATE INDEX idx_transactions_reports_object_path ON transactions_reports(object_path);

CREATE OR REPLACE FUNCTION update_transactions_reports_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_transactions_reports_updated_at
    BEFORE UPDATE ON transactions_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_transactions_reports_updated_at();

COMMENT ON TABLE transactions_reports IS 'Таблица для хранения информации о сгенерированных отчётах по транзакциям';
COMMENT ON COLUMN transactions_reports.id IS 'Уникальный идентификатор записи отчёта';
COMMENT ON COLUMN transactions_reports.object_path IS 'Путь к файлу отчёта в MinIO';
COMMENT ON COLUMN transactions_reports.comment IS 'Опциональный комментарий к отчёту';
COMMENT ON COLUMN transactions_reports.created_at IS 'Дата и время создания отчёта';
COMMENT ON COLUMN transactions_reports.updated_at IS 'Дата и время последнего обновления записи';