-- Up Migration: init_stock_position_movements
-- Type: tenant
-- Created: 2026-09-25

DO $$ BEGIN
    CREATE TYPE stock_movement_type AS ENUM (
        'write_off',   -- списание при отгрузке
        'return',      -- возврат (заготовка)
        'transfer',    -- перемещение (заготовка)
        'adjustment'   -- корректировка (заготовка)
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE stock_position_movements (
    outcome_batch_id uuid NOT NULL REFERENCES stock_batches(id) ON DELETE CASCADE,
    position_id uuid NOT NULL REFERENCES stock_positions(id) ON DELETE CASCADE,
    type stock_movement_type NOT NULL DEFAULT 'write_off',
    quantity numeric(15,3) NOT NULL CHECK (quantity > 0),
    comment text,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (outcome_batch_id, position_id)
);

CREATE INDEX idx_stock_movements_position_id ON stock_position_movements(position_id);
CREATE INDEX idx_stock_movements_outcome_batch_id ON stock_position_movements(outcome_batch_id);
CREATE INDEX idx_stock_movements_type ON stock_position_movements(type);
CREATE INDEX idx_stock_movements_created_at ON stock_position_movements(created_at DESC);

COMMENT ON TABLE stock_position_movements IS 'Движения позиций (списания, возвраты, перемещения)';
COMMENT ON COLUMN stock_position_movements.outcome_batch_id IS 'ID документа отгрузки/списания';
COMMENT ON COLUMN stock_position_movements.position_id IS 'ID позиции';
COMMENT ON COLUMN stock_position_movements.type IS 'Тип движения';
COMMENT ON COLUMN stock_position_movements.quantity IS 'Сколько списано в рамках этого движения';