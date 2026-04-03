-- Up Migration: add_isTech_to_support_ticket_messages
-- Type: public
-- Created: 2026-04-03 16:25:07

ALTER TABLE support_ticket_messages 
ADD COLUMN is_tech BOOLEAN NOT NULL DEFAULT false;