-- Up Migration: init_support_ticket_message_links
-- Type: public
-- Created: 2026-04-02 17:40:59

CREATE TABLE IF NOT EXISTS support_ticket_message_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES support_ticket_messages(id) ON DELETE CASCADE,
    link TEXT NOT NULL,
    capture TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP TRIGGER IF EXISTS update_support_ticket_message_links_updated_at ON support_ticket_message_links;

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_support_ticket_message_links_updated_at
    BEFORE UPDATE ON support_ticket_message_links
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();