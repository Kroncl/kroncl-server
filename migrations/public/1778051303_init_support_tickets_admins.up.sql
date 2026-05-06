-- Up Migration: init_support_tickets_admins
-- Type: public
-- Created: 2026-05-06 10:08:23

CREATE TABLE IF NOT EXISTS support_tickets_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT support_tickets_admins_ticket_id_unique UNIQUE (ticket_id)
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_admins_admin_id
    ON support_tickets_admins(admin_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_support_tickets_admins_updated_at
    BEFORE UPDATE ON support_tickets_admins
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();