-- Down Migration: confirmation_codes_replace_code_with_code_hash
-- Type: public
-- Created: 2026-04-13 04:16:33

ALTER TABLE confirmation_codes RENAME COLUMN code_hash TO code;
ALTER TABLE confirmation_codes ALTER COLUMN code TYPE VARCHAR(10);

DROP INDEX IF EXISTS idx_confirmation_codes_code_hash;
CREATE INDEX idx_confirmation_codes_code ON confirmation_codes(code);

CREATE OR REPLACE FUNCTION generate_confirmation_code(
    account_uuid UUID,
    code_type VARCHAR,
    code_length INTEGER DEFAULT 6,
    expiry_minutes INTEGER DEFAULT 5
) RETURNS VARCHAR AS $$
DECLARE
    new_code VARCHAR;
BEGIN
    new_code := LPAD(FLOOR(RANDOM() * POWER(10, code_length))::TEXT, code_length, '0');
    
    DELETE FROM confirmation_codes 
    WHERE account_id = account_uuid 
      AND type = code_type
      AND used = FALSE;
    
    INSERT INTO confirmation_codes (account_id, code, type, expires_at)
    VALUES (
        account_uuid,
        new_code,
        code_type,
        NOW() + (expiry_minutes || ' minutes')::INTERVAL
    );
    
    RETURN new_code;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION verify_confirmation_code(
    account_uuid UUID,
    input_code VARCHAR,
    code_type VARCHAR
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE confirmation_codes 
    SET used = TRUE
    WHERE account_id = account_uuid 
      AND code = input_code
      AND type = code_type
      AND used = FALSE
      AND expires_at > NOW();
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_active_confirmation_code(
    account_uuid UUID,
    code_type VARCHAR
) RETURNS TABLE (
    code VARCHAR,
    expires_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT c.code, c.expires_at
    FROM confirmation_codes c
    WHERE c.account_id = account_uuid
      AND c.type = code_type
      AND c.used = FALSE
      AND c.expires_at > NOW()
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;