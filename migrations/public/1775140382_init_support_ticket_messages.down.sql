-- Down Migration: init_support_ticket_messages
-- Type: public
-- Created: 2026-04-02 17:33:02

DROP TRIGGER IF EXISTS update_support_ticket_messages_updated_at ON support_ticket_messages;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS support_ticket_messages;