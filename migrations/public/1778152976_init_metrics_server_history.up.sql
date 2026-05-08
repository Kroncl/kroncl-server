-- Up Migration: init_metrics_server_history
-- Type: public
-- Created: 2026-05-07 14:22:56

CREATE TABLE IF NOT EXISTS metrics_server_history (
    id BIGSERIAL PRIMARY KEY,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- HTTP трафик
    requests_total INT NOT NULL DEFAULT 0,
    requests_5xx_total INT NOT NULL DEFAULT 0,
    requests_4xx_total INT NOT NULL DEFAULT 0,
    avg_response_time_ms INT NOT NULL DEFAULT 0,
    p95_response_time_ms INT NOT NULL DEFAULT 0,
    active_connections INT NOT NULL DEFAULT 0,
    
    -- Go runtime
    goroutines_count INT NOT NULL DEFAULT 0,
    heap_alloc_mb INT NOT NULL DEFAULT 0,
    heap_inuse_mb INT NOT NULL DEFAULT 0,
    gc_duration_ms INT NOT NULL DEFAULT 0,
    
    -- Воркеры
    db_worker_success BOOLEAN NOT NULL DEFAULT true,
    clientele_worker_success BOOLEAN NOT NULL DEFAULT true,
    
    -- Системные
    cpu_usage_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    memory_usage_mb INT NOT NULL DEFAULT 0,
    open_fds_count INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_metrics_server_recorded_at ON metrics_server_history(recorded_at);