-- Up Migration: init_metrics_media_history
-- Type: public
-- Created: 2026-05-24 15:09:23

CREATE TABLE IF NOT EXISTS metrics_media_history (
    id BIGSERIAL PRIMARY KEY,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Общая статистика по всем бакетам
    total_buckets INT NOT NULL DEFAULT 0,
    total_objects INT NOT NULL DEFAULT 0,
    total_size_mb BIGINT NOT NULL DEFAULT 0,
    
    -- Статистика по публичному бакету
    public_bucket_objects INT NOT NULL DEFAULT 0,
    public_bucket_size_mb BIGINT NOT NULL DEFAULT 0,
    
    -- Статистика по временному бакету
    temp_bucket_objects INT NOT NULL DEFAULT 0,
    temp_bucket_size_mb BIGINT NOT NULL DEFAULT 0,
    
    -- Статистика по арендным бакетам
    tenant_buckets_count INT NOT NULL DEFAULT 0,
    tenant_total_objects INT NOT NULL DEFAULT 0,
    tenant_total_size_mb BIGINT NOT NULL DEFAULT 0,
    avg_tenant_objects DECIMAL(10,2) NOT NULL DEFAULT 0,
    avg_tenant_size_mb DECIMAL(10,2) NOT NULL DEFAULT 0,
    
    -- Дополнительная аналитика
    largest_bucket_name VARCHAR(255),
    largest_bucket_objects INT NOT NULL DEFAULT 0,
    largest_bucket_size_mb BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metrics_media_recorded_at ON metrics_media_history(recorded_at);