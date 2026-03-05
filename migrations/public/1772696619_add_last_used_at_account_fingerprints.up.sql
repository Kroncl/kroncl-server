-- Up Migration: add_last_used_at_account_fingerprints
-- Type: public
-- Created: 2026-03-05 10:43:39

-- Add last_used_at column to account_fingerprints table
ALTER TABLE account_fingerprints 
ADD COLUMN last_used_at TIMESTAMP WITH TIME ZONE;

-- Add index for performance (опционально, но полезно)
CREATE INDEX idx_account_fingerprints_last_used 
ON account_fingerprints(last_used_at);

-- Comment for documentation
COMMENT ON COLUMN account_fingerprints.last_used_at 
IS 'Timestamp of last successful authentication with this fingerprint';