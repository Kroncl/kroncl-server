-- Migration: init_fingerprints
-- Description: Create fingerprints table for API key authentication
-- Type: public
-- Created: 2026-03-05

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum type for status (optional but recommended)
DO $$ BEGIN
    CREATE TYPE fingerprint_status AS ENUM ('active', 'inactive');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Create fingerprints table
CREATE TABLE fingerprints (
    -- Primary key using UUID v4
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Hashed fingerprint value (не сам ключ, а его хэш)
    -- Используем varchar с достаточной длиной для bcrypt/sha256
    hash VARCHAR(255) NOT NULL,
    
    -- Expiration timestamp (NULL means never expires)
    expired_at TIMESTAMP WITH TIME ZONE,
    
    -- Status using enum (recommended) or varchar
    status fingerprint_status NOT NULL DEFAULT 'active',
    -- Альтернатива без enum:
    -- status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Add constraint: hash must be unique
    CONSTRAINT fingerprints_hash_unique UNIQUE (hash)
);

-- Create indexes for better performance
CREATE INDEX idx_fingerprints_status ON fingerprints(status);
CREATE INDEX idx_fingerprints_expired_at ON fingerprints(expired_at);
CREATE INDEX idx_fingerprints_created_at ON fingerprints(created_at);

-- Composite index for common queries (active + not expired)
CREATE INDEX idx_fingerprints_active_not_expired 
    ON fingerprints(status, expired_at) 
    WHERE status = 'active';

-- Optional: Add comment on table for documentation
COMMENT ON TABLE fingerprints IS 'Stores hashed fingerprints for API key authentication';
COMMENT ON COLUMN fingerprints.hash IS 'Bcrypt or SHA256 hash of the fingerprint key';
COMMENT ON COLUMN fingerprints.status IS 'active - ключ работает, inactive - отозван/заблокирован';
COMMENT ON COLUMN fingerprints.expired_at IS 'NULL = бессрочный, иначе дата истечения';