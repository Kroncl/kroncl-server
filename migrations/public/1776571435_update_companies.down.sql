-- Down Migration: update_companies
-- Type: public
-- Created: 2026-04-19 07:03:55

ALTER TABLE companies DROP COLUMN email;
ALTER TABLE companies DROP COLUMN region;
ALTER TABLE companies DROP COLUMN site;
ALTER TABLE companies DROP COLUMN metadata;