-- Down Migration: init_catalog_unit_pivot_table
-- Type: tenant
-- Created: 2026-02-28 08:53:54

-- Drop trigger
DROP TRIGGER IF EXISTS update_catalog_unit_category_updated_at ON catalog_unit_category;

-- Drop function
DROP FUNCTION IF EXISTS update_catalog_unit_category_updated_at();

-- Drop table
DROP TABLE IF EXISTS catalog_unit_category;