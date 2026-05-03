package coreworkers

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) CollectMetrics(ctx context.Context) (*MetricsDBSnapshot, error) {
	var stats MetricsDBSnapshot

	query := `
		SELECT 
			COALESCE(pg_database_size(current_database()) / 1024 / 1024, 0) as total_size_mb,
			(SELECT COUNT(*) FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname != 'information_schema') as total_schemas,
			(SELECT COUNT(*) FROM pg_namespace WHERE nspname LIKE 'company_%') as company_schemas,
			(SELECT SUM(reltuples::bigint) FROM pg_class WHERE relkind = 'r') as total_tables,
			(SELECT COUNT(*) FROM pg_index) as total_indexes,
			(SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active') as active_conns,
			COALESCE(xact_commit, 0) as xact_commit,
			COALESCE(xact_rollback, 0) as xact_rollback,
			COALESCE(tup_returned, 0) as tup_returned,
			COALESCE(tup_fetched, 0) as tup_fetched,
			COALESCE(tup_inserted, 0) as tup_inserted,
			COALESCE(tup_updated, 0) as tup_updated,
			COALESCE(tup_deleted, 0) as tup_deleted,
			COALESCE(blks_read, 0) as blks_read,
			COALESCE(blks_hit, 0) as blks_hit
		FROM pg_stat_database
		WHERE datname = current_database()
	`

	err := s.pool.QueryRow(ctx, query).Scan(
		&stats.TotalDatabaseSizeMB,
		&stats.TotalSchemasCount,
		&stats.CompanySchemasCount,
		&stats.TotalTablesCount,
		&stats.TotalIndexesCount,
		&stats.TotalActiveConnections,
		&stats.XactCommit,
		&stats.XactRollback,
		&stats.TupReturned,
		&stats.TupFetched,
		&stats.TupInserted,
		&stats.TupUpdated,
		&stats.TupDeleted,
		&stats.BlksRead,
		&stats.BlksHit,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to collect metrics: %w", err)
	}

	stats.RecordedAt = time.Now()
	return &stats, nil
}

func (s *Service) SaveMetricsSnapshot(ctx context.Context, stats *MetricsDBSnapshot) error {
	query := `
		INSERT INTO metrics_db_history (
			recorded_at, total_database_size_mb, total_schemas_count, company_schemas_count,
			total_tables_count, total_indexes_count, total_active_connections,
			xact_commit, xact_rollback, tup_returned, tup_fetched,
			tup_inserted, tup_updated, tup_deleted, blks_read, blks_hit
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err := s.pool.Exec(ctx, query,
		stats.RecordedAt,
		stats.TotalDatabaseSizeMB,
		stats.TotalSchemasCount,
		stats.CompanySchemasCount,
		stats.TotalTablesCount,
		stats.TotalIndexesCount,
		stats.TotalActiveConnections,
		stats.XactCommit,
		stats.XactRollback,
		stats.TupReturned,
		stats.TupFetched,
		stats.TupInserted,
		stats.TupUpdated,
		stats.TupDeleted,
		stats.BlksRead,
		stats.BlksHit,
	)

	if err != nil {
		return fmt.Errorf("failed to save metrics snapshot: %w", err)
	}

	return nil
}

// GetMetricsHistory возвращает историю метрик с фильтрацией
func (s *Service) GetMetricsHistory(ctx context.Context, startDate, endDate *time.Time, limit int) ([]MetricsDBSnapshot, error) {
	query := `
		SELECT 
			recorded_at, total_database_size_mb, total_schemas_count, company_schemas_count,
			total_tables_count, total_indexes_count, total_active_connections,
			xact_commit, xact_rollback, tup_returned, tup_fetched,
			tup_inserted, tup_updated, tup_deleted, blks_read, blks_hit
		FROM metrics_db_history
		WHERE 1=1
	`

	args := []interface{}{}
	argCounter := 1

	if startDate != nil {
		query += fmt.Sprintf(" AND recorded_at >= $%d", argCounter)
		args = append(args, *startDate)
		argCounter++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND recorded_at <= $%d", argCounter)
		args = append(args, *endDate)
		argCounter++
	}

	query += " ORDER BY recorded_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCounter)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics history: %w", err)
	}
	defer rows.Close()

	var metrics []MetricsDBSnapshot
	for rows.Next() {
		var m MetricsDBSnapshot
		err := rows.Scan(
			&m.RecordedAt,
			&m.TotalDatabaseSizeMB,
			&m.TotalSchemasCount,
			&m.CompanySchemasCount,
			&m.TotalTablesCount,
			&m.TotalIndexesCount,
			&m.TotalActiveConnections,
			&m.XactCommit,
			&m.XactRollback,
			&m.TupReturned,
			&m.TupFetched,
			&m.TupInserted,
			&m.TupUpdated,
			&m.TupDeleted,
			&m.BlksRead,
			&m.BlksHit,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}
