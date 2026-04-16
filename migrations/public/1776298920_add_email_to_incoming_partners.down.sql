-- Down Migration: add_email_to_incoming_partners
-- Type: public
-- Created: 2026-04-16 03:22:00

ALTER TABLE incoming_partners DROP COLUMN email;