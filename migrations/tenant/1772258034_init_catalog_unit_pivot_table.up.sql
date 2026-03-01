-- Up Migration: init_catalog_unit_pivot_table
-- Type: tenant
-- Created: 2026-02-28 08:53:54

-- Catalog unit-category pivot table
CREATE TABLE catalog_unit_category (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id uuid NOT NULL REFERENCES catalog_units(id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES catalog_categories(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(unit_id)
);

-- Indexes
CREATE INDEX idx_catalog_unit_category_unit_id ON catalog_unit_category(unit_id);
CREATE INDEX idx_catalog_unit_category_category_id ON catalog_unit_category(category_id);
CREATE INDEX idx_catalog_unit_category_created_at ON catalog_unit_category(created_at DESC);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_catalog_unit_category_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_catalog_unit_category_updated_at
    BEFORE UPDATE ON catalog_unit_category
    FOR EACH ROW
    EXECUTE FUNCTION update_catalog_unit_category_updated_at();

-- Comments
COMMENT ON COLUMN catalog_unit_category.unit_id IS 'Reference to catalog unit';
COMMENT ON COLUMN catalog_unit_category.category_id IS 'Reference to catalog category';