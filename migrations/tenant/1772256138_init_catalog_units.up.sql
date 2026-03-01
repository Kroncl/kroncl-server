-- Up Migration: init_catalog_units
-- Type: tenant
-- Created: 2026-02-28 08:22:18

-- Unit type (product/service)
DO $$ BEGIN
    CREATE TYPE catalog_unit_type AS ENUM ('product', 'service');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Unit status
DO $$ BEGIN
    CREATE TYPE catalog_unit_status AS ENUM ('active', 'inactive');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Inventory tracking type
DO $$ BEGIN
    CREATE TYPE inventory_type AS ENUM (
        'tracked',      -- считаем остатки (обычный товар)
        'untracked'     -- не считаем (услуга, цифровой товар, лицензия)
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Tracked type (FIFO/LIFO)
DO $$ BEGIN
    CREATE TYPE tracked_type AS ENUM ('fifo', 'lifo');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Currency
DO $$ BEGIN
    CREATE TYPE currency_type AS ENUM ('RUB', 'USD', 'EUR', 'KZT');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Catalog units table (products and services)
CREATE TABLE catalog_units (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    comment text,
    type catalog_unit_type NOT NULL,
    status catalog_unit_status NOT NULL DEFAULT 'active',
    inventory_type inventory_type NOT NULL DEFAULT 'tracked',
    tracked_type tracked_type,  -- только для tracked, может быть NULL для untracked
    unit varchar(20) NOT NULL DEFAULT 'pcs',
    
    -- Цены
    sale_price numeric(15,2) NOT NULL CHECK (sale_price >= 0),
    purchase_price numeric(15,2),  -- только для tracked, может быть NULL для untracked
    currency currency_type NOT NULL DEFAULT 'RUB',
    
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_catalog_units_type ON catalog_units(type);
CREATE INDEX idx_catalog_units_status ON catalog_units(status);
CREATE INDEX idx_catalog_units_inventory_type ON catalog_units(inventory_type);
CREATE INDEX idx_catalog_units_tracked_type ON catalog_units(tracked_type);
CREATE INDEX idx_catalog_units_name ON catalog_units(name);
CREATE INDEX idx_catalog_units_sale_price ON catalog_units(sale_price);
CREATE INDEX idx_catalog_units_purchase_price ON catalog_units(purchase_price);
CREATE INDEX idx_catalog_units_created_at ON catalog_units(created_at DESC);

-- CHECK constraints
-- 1. Услуги не могут быть tracked
ALTER TABLE catalog_units ADD CONSTRAINT service_cannot_be_tracked 
    CHECK (
        NOT (type = 'service' AND inventory_type = 'tracked')
    );

-- 2. tracked_type обязателен для tracked, запрещён для untracked
ALTER TABLE catalog_units ADD CONSTRAINT tracked_type_required_for_tracked
    CHECK (
        (inventory_type = 'tracked' AND tracked_type IS NOT NULL) OR
        (inventory_type = 'untracked' AND tracked_type IS NULL)
    );

-- 3. purchase_price обязателен для tracked, запрещён для untracked
ALTER TABLE catalog_units ADD CONSTRAINT purchase_price_required_for_tracked
    CHECK (
        (inventory_type = 'tracked' AND purchase_price IS NOT NULL AND purchase_price >= 0) OR
        (inventory_type = 'untracked' AND purchase_price IS NULL)
    );

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_catalog_units_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_catalog_units_updated_at
    BEFORE UPDATE ON catalog_units
    FOR EACH ROW
    EXECUTE FUNCTION update_catalog_units_updated_at();

-- Comments
COMMENT ON TABLE catalog_units IS 'Catalog units (products and services)';
COMMENT ON COLUMN catalog_units.name IS 'Unit name';
COMMENT ON COLUMN catalog_units.comment IS 'Description/comment';
COMMENT ON COLUMN catalog_units.type IS 'product or service';
COMMENT ON COLUMN catalog_units.status IS 'active or inactive';
COMMENT ON COLUMN catalog_units.inventory_type IS 'tracked - count inventory, untracked - infinite (services, digital goods)';
COMMENT ON COLUMN catalog_units.tracked_type IS 'FIFO or LIFO - required for tracked items';
COMMENT ON COLUMN catalog_units.unit IS 'Measurement unit (pcs, kg, l, etc)';
COMMENT ON COLUMN catalog_units.sale_price IS 'Selling price (for all)';
COMMENT ON COLUMN catalog_units.purchase_price IS 'Purchase price - required for tracked items only';
COMMENT ON COLUMN catalog_units.currency IS 'Price currency';
COMMENT ON COLUMN catalog_units.metadata IS 'Additional data (characteristics, etc)';