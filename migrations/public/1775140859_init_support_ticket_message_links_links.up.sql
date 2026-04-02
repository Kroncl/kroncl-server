-- Up Migration: init_support_ticket_message_links
-- Type: public
-- Created: 2026-04-02 17:40:59

DROP TRIGGER IF EXISTS update_support_ticket_message_links_updated_at ON support_ticket_message_links;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS support_ticket_message_links;