-- Down Migration: drop_stock_position_batch

CREATE TABLE stock_position_batch (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    position_id uuid NOT NULL REFERENCES stock_positions(id) ON DELETE CASCADE,
    batch_id uuid NOT NULL REFERENCES stock_batches(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(position_id, batch_id)
);

CREATE INDEX idx_stock_position_batch_position_id ON stock_position_batch(position_id);
CREATE INDEX idx_stock_position_batch_batch_id ON stock_position_batch(batch_id);
CREATE INDEX idx_stock_position_batch_created_at ON stock_position_batch(created_at DESC);

-- Восстанавливаем данные из stock_positions.income_batch_id
INSERT INTO stock_position_batch (position_id, batch_id)
SELECT id, income_batch_id FROM stock_positions
WHERE income_batch_id IS NOT NULL
ON CONFLICT DO NOTHING;

COMMENT ON TABLE stock_position_batch IS 'Связь между позициями на складе и партиями движения';