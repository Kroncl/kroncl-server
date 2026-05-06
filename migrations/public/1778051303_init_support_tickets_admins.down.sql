-- Down Migration: init_support_tickets_admins
-- Type: public
-- Created: 2026-05-06 10:08:23

DROP TRIGGER IF EXISTS update_support_tickets_admins_updated_at ON support_tickets_admins;
DROP TABLE IF EXISTS support_tickets_admins;