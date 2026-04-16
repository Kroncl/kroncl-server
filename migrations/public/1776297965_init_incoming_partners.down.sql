-- Down Migration: init_incoming_partners
-- Type: public
-- Created: 2026-04-16 03:06:05

DROP TABLE IF EXISTS incoming_partners;
DROP TYPE IF EXISTS incoming_partner_status;
DROP TYPE IF EXISTS incoming_partner_type;