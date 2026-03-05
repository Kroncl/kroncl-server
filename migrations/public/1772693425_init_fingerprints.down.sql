-- Migration rollback: init_fingerprints
-- Description: Drop fingerprints table and related objects

-- Drop the table first (cascade in case there are dependencies)
DROP TABLE IF EXISTS fingerprints CASCADE;

-- Drop the enum type (only if you created it)
DROP TYPE IF EXISTS fingerprint_status;

-- Note: We don't drop uuid-ossp extension as it might be used elsewhere
-- If you're sure it's safe to remove:
-- DROP EXTENSION IF EXISTS "uuid-ossp";