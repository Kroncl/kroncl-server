-- Up Migration: init_currency_rates
-- Type: public
-- Created: 2026-06-21 07:16:08

CREATE TABLE currencies (
    id      VARCHAR(10) PRIMARY KEY,  -- USD, EUR, BTC, TON
    name    VARCHAR(100) NOT NULL,     -- US Dollar, Euro, Bitcoin
    type    VARCHAR(20) NOT NULL DEFAULT 'fiat',  -- fiat, crypto
    symbol  VARCHAR(10)               -- $, €, ₿
);

INSERT INTO currencies (id, name, type, symbol) VALUES
    ('USD', 'Доллар США', 'fiat', '$'),
    ('EUR', 'Евро', 'fiat', '€'),
    ('KZT', 'Казахстанский тенге', 'fiat', '₸'),
    ('CNY', 'Китайский юань', 'fiat', '¥'),
    ('AMD', 'Армянский драм', 'fiat', '֏'),
    ('GBP', 'Британский фунт', 'fiat', '£'),
    ('TRY', 'Турецкая лира', 'fiat', '₺'),
    ('AED', 'Дирхам ОАЭ', 'fiat', 'د.إ'),
    ('BYN', 'Белорусский рубль', 'fiat', 'Br'),
    ('BTC', 'Биткоин', 'crypto', '₿'),
    ('ETH', 'Эфириум', 'crypto', 'Ξ'),
    ('SOL', 'Солана', 'crypto', '◎'),
    ('USDT', 'Тезер', 'crypto', '₮'),
    ('USDC', 'USD Coin', 'crypto', ''),
    ('TON', 'The Open Network', 'crypto', '')
ON CONFLICT DO NOTHING;

CREATE TABLE currency_rates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    currency_id VARCHAR(10) NOT NULL REFERENCES currencies(id),
    rate        NUMERIC(18, 8) NOT NULL,  -- сколько RUB за 1 единицу валюты
    source      VARCHAR(50) NOT NULL DEFAULT 'manual',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_currency_rates_currency_id ON currency_rates (currency_id, updated_at DESC);
CREATE INDEX idx_currency_rates_source ON currency_rates (source);