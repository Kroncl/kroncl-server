-- Down Migration: init_catalog_units
-- Type: tenant
-- Created: 2026-02-28 08:22:18

-- Drop CHECK constraints
ALTER TABLE catalog_units DROP CONSTRAINT IF EXISTS service_cannot_be_tracked;
ALTER TABLE catalog_units DROP CONSTRAINT IF EXISTS tracked_type_required_for_tracked;
ALTER TABLE catalog_units DROP CONSTRAINT IF EXISTS purchase_price_required_for_tracked;

-- Drop trigger
DROP TRIGGER IF EXISTS update_catalog_units_updated_at ON catalog_units;

-- Drop function
DROP FUNCTION IF EXISTS update_catalog_units_updated_at();

-- Drop table
DROP TABLE IF EXISTS catalog_units;

-- Drop enum types
DROP TYPE IF EXISTS catalog_unit_type;
DROP TYPE IF EXISTS catalog_unit_status;
DROP TYPE IF EXISTS inventory_type;
DROP TYPE IF EXISTS tracked_type;
DROP TYPE IF EXISTS currency_type;