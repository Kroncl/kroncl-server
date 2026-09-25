-- Up Migration: init_barcodes
-- Type: tenant
-- Created: 2026-09-25

CREATE TABLE barcodes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    barcode varchar(255) UNIQUE NOT NULL,
    maker varchar(255),
    catalog_unit_id uuid REFERENCES catalog_units(id) ON DELETE SET NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_barcodes_barcode ON barcodes(barcode);
CREATE INDEX idx_barcodes_catalog_unit_id ON barcodes(catalog_unit_id);
CREATE INDEX idx_barcodes_created_at ON barcodes(created_at DESC);

CREATE OR REPLACE FUNCTION update_barcodes_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_barcodes_updated_at
    BEFORE UPDATE ON barcodes
    FOR EACH ROW
    EXECUTE FUNCTION update_barcodes_updated_at();

COMMENT ON TABLE barcodes IS 'Словарь штрихкодов';
COMMENT ON COLUMN barcodes.barcode IS 'Штрихкод (уникальный)';
COMMENT ON COLUMN barcodes.maker IS 'Производитель';
COMMENT ON COLUMN barcodes.catalog_unit_id IS 'Опциональная привязка к товарной позиции каталога';