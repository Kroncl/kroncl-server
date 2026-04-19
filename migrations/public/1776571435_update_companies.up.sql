-- Up Migration: update_companies
-- Type: public
-- Created: 2026-04-19 07:03:55

ALTER TABLE companies ADD COLUMN email VARCHAR(255);
ALTER TABLE companies ADD COLUMN region VARCHAR(10) NOT NULL DEFAULT 'ru-RU';
ALTER TABLE companies ADD COLUMN site VARCHAR(255);
ALTER TABLE companies ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}'::jsonb;