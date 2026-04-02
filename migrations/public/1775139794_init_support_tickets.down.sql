-- Down Migration: init_support_tickets
-- Type: public
-- Created: 2026-04-02 17:23:14

DROP TRIGGER IF EXISTS update_support_tickets_updated_at ON support_tickets;
DROP TABLE IF EXISTS support_tickets;
DROP TYPE IF EXISTS support_ticket_status;