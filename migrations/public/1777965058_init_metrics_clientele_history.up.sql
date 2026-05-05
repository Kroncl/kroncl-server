-- Up Migration: init_metrics_clientele_history
-- Type: public
-- Created: 2026-05-05 10:10:58

CREATE TABLE IF NOT EXISTS metrics_clientele_history (
    id BIGSERIAL PRIMARY KEY,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Аккаунты
    total_accounts INT NOT NULL DEFAULT 0,
    confirmed_accounts INT NOT NULL DEFAULT 0,
    waiting_accounts INT NOT NULL DEFAULT 0,
    admin_accounts INT NOT NULL DEFAULT 0,
    
    -- Статистика по типам аккаунтов
    account_type_owner INT NOT NULL DEFAULT 0,
    account_type_employee INT NOT NULL DEFAULT 0,
    account_type_admin INT NOT NULL DEFAULT 0,
    account_type_outsourcing INT NOT NULL DEFAULT 0,
    account_type_tech INT NOT NULL DEFAULT 0,
    
    -- Компании
    total_companies INT NOT NULL DEFAULT 0,
    public_companies INT NOT NULL DEFAULT 0,
    private_companies INT NOT NULL DEFAULT 0,
    
    -- Связи
    total_company_accounts INT NOT NULL DEFAULT 0,
    avg_accounts_per_company DECIMAL(10,2) NOT NULL DEFAULT 0,
    max_accounts_in_company INT NOT NULL DEFAULT 0,
    
    -- Транзакции
    total_transactions INT NOT NULL DEFAULT 0,
    success_transactions INT NOT NULL DEFAULT 0,
    pending_transactions INT NOT NULL DEFAULT 0,
    trial_transactions INT NOT NULL DEFAULT 0,
    
    -- Активность
    active_companies_7d INT NOT NULL DEFAULT 0,
    active_companies_30d INT NOT NULL DEFAULT 0,
    
    -- Схемы
    company_schemas_without_data INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_metrics_clientele_recorded_at ON metrics_clientele_history(recorded_at);