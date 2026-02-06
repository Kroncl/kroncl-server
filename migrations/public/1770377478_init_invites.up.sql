-- Up Migration: init_invites
-- Type: public
-- Created: 2026-02-06 14:31:18

-- Создание таблицы приглашений в компании
CREATE TABLE company_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    company_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Ограничения
    CONSTRAINT fk_company_invitations_company
        FOREIGN KEY (company_id)
        REFERENCES companies(id)
        ON DELETE CASCADE,
    
    CONSTRAINT chk_company_invitations_status
        CHECK (status IN ('waiting', 'rejected', 'accepted')),
    
    CONSTRAINT uq_company_invitations_email_company
        UNIQUE (email, company_id)
);

-- Создание индексов для оптимизации запросов
CREATE INDEX idx_company_invitations_email ON company_invitations(email);
CREATE INDEX idx_company_invitations_company_id ON company_invitations(company_id);
CREATE INDEX idx_company_invitations_status ON company_invitations(status);
CREATE INDEX idx_company_invitations_created_at ON company_invitations(created_at DESC);

-- Создание триггера для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_company_invitations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_company_invitations_updated_at
    BEFORE UPDATE ON company_invitations
    FOR EACH ROW
    EXECUTE FUNCTION update_company_invitations_updated_at();

-- Добавление комментариев к таблице и колонкам
COMMENT ON TABLE company_invitations IS 'Таблица приглашений пользователей в компании';
COMMENT ON COLUMN company_invitations.id IS 'Уникальный идентификатор приглашения';
COMMENT ON COLUMN company_invitations.email IS 'Email приглашенного пользователя';
COMMENT ON COLUMN company_invitations.company_id IS 'ID компании, в которую приглашают';
COMMENT ON COLUMN company_invitations.status IS 'Статус приглашения: waiting, rejected, accepted';
COMMENT ON COLUMN company_invitations.created_at IS 'Дата и время создания приглашения';
COMMENT ON COLUMN company_invitations.updated_at IS 'Дата и время последнего обновления приглашения';