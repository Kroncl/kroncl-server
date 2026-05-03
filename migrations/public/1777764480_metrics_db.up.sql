-- Up Migration: metrics_db
-- Type: public
-- Created: 2026-05-03 02:28:00

CREATE TABLE IF NOT EXISTS metrics_db_history (
    id BIGSERIAL PRIMARY KEY,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- общая статистика БД
    total_database_size_mb BIGINT NOT NULL DEFAULT 0,
    total_schemas_count INT NOT NULL DEFAULT 0,
    company_schemas_count INT NOT NULL DEFAULT 0,
    total_tables_count INT NOT NULL DEFAULT 0,
    total_indexes_count INT NOT NULL DEFAULT 0,
    total_active_connections INT NOT NULL DEFAULT 0,
    
    -- нагрузка из pg_stat_database
    xact_commit BIGINT NOT NULL DEFAULT 0,
    xact_rollback BIGINT NOT NULL DEFAULT 0,
    tup_returned BIGINT NOT NULL DEFAULT 0,
    tup_fetched BIGINT NOT NULL DEFAULT 0,
    tup_inserted BIGINT NOT NULL DEFAULT 0,
    tup_updated BIGINT NOT NULL DEFAULT 0,
    tup_deleted BIGINT NOT NULL DEFAULT 0,
    blks_read BIGINT NOT NULL DEFAULT 0,
    blks_hit BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_metrics_db_recorded_at ON metrics_db_history(recorded_at);