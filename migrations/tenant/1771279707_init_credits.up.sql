-- Up Migration: init_credits
-- Type: tenant
-- Created: 2026-02-17 01:08:27

-- Типы для кредитов
DO $$ BEGIN
    CREATE TYPE credit_type AS ENUM ('debt', 'credit');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE credit_status AS ENUM ('active', 'closed');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Основная таблица кредитов
CREATE TABLE credits (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    comment text,
    type credit_type NOT NULL,
    status credit_status NOT NULL DEFAULT 'active',
    total_amount bigint NOT NULL CHECK (total_amount > 0),
    currency currency_type NOT NULL DEFAULT 'RUB',
    interest_rate numeric(5,2) DEFAULT 0 CHECK (interest_rate >= 0),
    start_date date NOT NULL,
    end_date date NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    -- Валидация дат
    CONSTRAINT credits_dates_check CHECK (end_date >= start_date)
);

-- Индексы
CREATE INDEX idx_credits_type ON credits(type);
CREATE INDEX idx_credits_status ON credits(status);
CREATE INDEX idx_credits_start_date ON credits(start_date);
CREATE INDEX idx_credits_end_date ON credits(end_date);
CREATE INDEX idx_credits_created_at ON credits(created_at DESC);
CREATE INDEX idx_credits_currency ON credits(currency);

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_credits_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_credits_updated_at
    BEFORE UPDATE ON credits
    FOR EACH ROW
    EXECUTE FUNCTION update_credits_updated_at();

-- Комментарии
COMMENT ON TABLE credits IS 'Кредиты и займы (дебиторская/кредиторская задолженность)';
COMMENT ON COLUMN credits.type IS 'debt - мы должны, credit - нам должны';
COMMENT ON COLUMN credits.total_amount IS 'Общая сумма кредита/займа';
COMMENT ON COLUMN credits.interest_rate IS 'Процентная ставка (годовых)';
COMMENT ON COLUMN credits.metadata IS 'Дополнительные параметры (график платежей, штрафы и т.д.)';