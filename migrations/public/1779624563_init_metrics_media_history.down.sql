-- Down Migration: init_metrics_media_history
-- Type: public
-- Created: 2026-05-24 15:09:23

DROP INDEX IF EXISTS idx_metrics_media_recorded_at;
DROP TABLE IF EXISTS metrics_media_history;