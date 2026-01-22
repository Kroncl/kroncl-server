package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetStorageSources(ctx context.Context, schemaName string) (*StorageSources, error) {
	query := `
		-- Проверяем существование схемы
		WITH schema_check AS (
			SELECT EXISTS(
				SELECT 1 FROM information_schema.schemata 
				WHERE schema_name = $1
			) as schema_exists
		),
		-- Статистика по таблицам
		table_stats AS (
			SELECT 
				COUNT(*) as table_count,
				COALESCE(SUM(n_live_tup), 0) as total_rows,
				COALESCE(SUM(n_dead_tup), 0) as dead_rows,
				MAX(last_vacuum) as last_vacuum_time,
				MAX(last_autovacuum) as last_autovacuum_time,
				MAX(last_analyze) as last_analyze_time
			FROM pg_stat_user_tables 
			WHERE schemaname = $1
		),
		-- Размеры объектов
		size_stats AS (
			SELECT 
				COALESCE(SUM(pg_table_size(quote_ident(schemaname) || '.' || quote_ident(tablename))), 0) as table_bytes,
				COALESCE(SUM(pg_indexes_size(quote_ident(schemaname) || '.' || quote_ident(tablename))), 0) as index_bytes,
				COALESCE(SUM(pg_total_relation_size(quote_ident(schemaname) || '.' || quote_ident(tablename))), 0) as total_bytes
			FROM pg_tables 
			WHERE schemaname = $1
		),
		-- TOAST размеры
		toast_stats AS (
			SELECT COALESCE(SUM(pg_total_relation_size(quote_ident(schemaname) || '.' || quote_ident(tablename))), 0) as toast_bytes
			FROM pg_tables 
			WHERE schemaname = $1 
				AND tablename LIKE 'pg_toast_%'
		),
		-- Количество индексов
		index_stats AS (
			SELECT COUNT(*) as index_count
			FROM pg_indexes 
			WHERE schemaname = $1
		),
		-- Количество последовательностей
		sequence_stats AS (
			SELECT COUNT(*) as sequence_count
			FROM information_schema.sequences 
			WHERE sequence_schema = $1
		),
		-- Количество представлений
		view_stats AS (
			SELECT 
				COUNT(*) filter (WHERE table_type = 'VIEW') as view_count,
				COUNT(*) filter (WHERE table_type = 'MATERIALIZED VIEW') as materialized_view_count
			FROM information_schema.tables 
			WHERE table_schema = $1
		),
		-- Активные соединения
		connection_stats AS (
			SELECT COUNT(*) as active_connections
			FROM pg_stat_activity 
			WHERE state = 'active' 
				AND query NOT LIKE '%pg_stat_activity%'
				AND datname = current_database()
		)
		SELECT 
			sc.schema_exists,
			ts.table_count,
			ts.total_rows,
			ts.dead_rows,
			ss.table_bytes,
			ss.index_bytes,
			ss.total_bytes,
			tst.toast_bytes,
			isx.index_count,
			vw.view_count,
			vw.materialized_view_count,
			seq.sequence_count,
			conn.active_connections,
			ts.last_vacuum_time,
			ts.last_autovacuum_time,
			ts.last_analyze_time
		FROM schema_check sc
		CROSS JOIN table_stats ts
		CROSS JOIN size_stats ss
		CROSS JOIN toast_stats tst
		CROSS JOIN index_stats isx
		CROSS JOIN sequence_stats seq
		CROSS JOIN view_stats vw
		CROSS JOIN connection_stats conn
	`

	var sources StorageSources
	var (
		schemaExists                                        bool
		tableBytes, indexBytes, totalBytes, toastBytes      int64
		lastVacuumTime, lastAutovacuumTime, lastAnalyzeTime *time.Time
	)

	err := r.pool.QueryRow(ctx, query, schemaName).Scan(
		&schemaExists,
		&sources.TableCount,
		&sources.TotalRows,
		&sources.DeadRows,
		&tableBytes,
		&indexBytes,
		&totalBytes,
		&toastBytes,
		&sources.IndexCount,
		&sources.ViewCount,
		&sources.MaterializedViewCount,
		&sources.SequenceCount,
		&sources.ActiveConnections,
		&lastVacuumTime,
		&lastAutovacuumTime,
		&lastAnalyzeTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage sources for schema %s: %w", schemaName, err)
	}

	// Заполняем основную структуру
	sources.SchemaName = schemaName
	sources.SchemaExists = schemaExists

	// Размеры в MB
	sources.TotalSizeMB = float64(totalBytes) / (1024 * 1024)
	sources.TableSizeMB = float64(tableBytes) / (1024 * 1024)
	sources.IndexSizeMB = float64(indexBytes) / (1024 * 1024)
	sources.ToastSizeMB = float64(toastBytes) / (1024 * 1024)

	// Форматируем общий размер для человека
	sources.TotalSizePretty = formatBytes(totalBytes)

	// Форматируем временные метки
	sources.LastVacuum = formatTime(lastVacuumTime)
	sources.LastAutovacuum = formatTime(lastAutovacuumTime)
	sources.LastAnalyze = formatTime(lastAnalyzeTime)

	// Время обновления статистики
	now := time.Now().Format(time.RFC3339)
	sources.UpdatedAt = &now

	return &sources, nil
}

func (r *Repository) CreateStorageRecord(ctx context.Context, companyID string) (*Storage, error) {
	query := `
		INSERT INTO company_storage (company_id, schema_name, status)
		VALUES ($1, generate_tenant_schema_name($1), 'provisioning')
		RETURNING id, company_id, schema_name, status, storage_type, metadata, created_at, updated_at
	`

	var storage Storage
	err := r.pool.QueryRow(ctx, query, companyID).Scan(
		&storage.ID,
		&storage.CompanyID,
		&storage.SchemaName,
		&storage.Status,
		&storage.StorageType,
		&storage.Metadata,
		&storage.CreatedAt,
		&storage.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create storage record: %w", err)
	}

	return &storage, nil
}

func (r *Repository) GetStorageStatus(ctx context.Context, storageID string) (string, error) {
	query := `SELECT status FROM company_storage WHERE id = $1`

	var status string
	err := r.pool.QueryRow(ctx, query, storageID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("storage not found: %s", storageID)
		}
		return "", fmt.Errorf("failed to get storage status: %w", err)
	}

	return status, nil
}

func (r *Repository) UpdateStorageStatus(ctx context.Context, storageID, status string) error {
	query := `
		UPDATE company_storage 
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, status, storageID)
	if err != nil {
		return fmt.Errorf("failed to update storage status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	return nil
}

func (r *Repository) GetStorageByCompanyID(ctx context.Context, companyID string) (*Storage, error) {
	query := `
		SELECT id, company_id, schema_name, status, storage_type, metadata, created_at, updated_at
		FROM company_storage
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var storage Storage
	err := r.pool.QueryRow(ctx, query, companyID).Scan(
		&storage.ID,
		&storage.CompanyID,
		&storage.SchemaName,
		&storage.Status,
		&storage.StorageType,
		&storage.Metadata,
		&storage.CreatedAt,
		&storage.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get storage for company %s: %w", companyID, err)
	}

	return &storage, nil
}
