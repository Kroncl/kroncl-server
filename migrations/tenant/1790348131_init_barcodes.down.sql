-- Down Migration: init_barcodes

DROP TRIGGER IF EXISTS update_barcodes_updated_at ON barcodes;
DROP FUNCTION IF EXISTS update_barcodes_updated_at();

DROP TABLE IF EXISTS barcodes;