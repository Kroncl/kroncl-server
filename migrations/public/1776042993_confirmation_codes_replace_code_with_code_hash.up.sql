-- Up Migration: confirmation_codes_replace_code_with_code_hash
-- Type: public
-- Created: 2026-04-13 04:16:33

ALTER TABLE confirmation_codes RENAME COLUMN code TO code_hash;
ALTER TABLE confirmation_codes ALTER COLUMN code_hash TYPE VARCHAR(255);

DROP INDEX IF EXISTS idx_confirmation_codes_code;
CREATE INDEX idx_confirmation_codes_code_hash ON confirmation_codes(code_hash);

DROP FUNCTION IF EXISTS generate_confirmation_code(UUID, VARCHAR, INTEGER, INTEGER);
DROP FUNCTION IF EXISTS verify_confirmation_code(UUID, VARCHAR, VARCHAR);
DROP FUNCTION IF EXISTS get_active_confirmation_code(UUID, VARCHAR);