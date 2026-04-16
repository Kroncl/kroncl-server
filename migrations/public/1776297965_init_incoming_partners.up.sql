-- Up Migration: init_incoming_partners
-- Type: public
-- Created: 2026-04-16 03:06:05

CREATE TYPE incoming_partner_type AS ENUM ('public', 'private');
CREATE TYPE incoming_partner_status AS ENUM ('success', 'waiting', 'banned');

CREATE TABLE incoming_partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type incoming_partner_type NOT NULL,
    text TEXT,
    status incoming_partner_status NOT NULL DEFAULT 'waiting',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);