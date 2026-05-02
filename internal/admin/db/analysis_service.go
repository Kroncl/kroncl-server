package admindb

import (
	"context"
	"fmt"
)

func (s *Service) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	var stats SystemStats

	query := `
		SELECT 
			(SELECT pg_database_size(current_database()) / 1024 / 1024) as total_size_mb,
			(SELECT COUNT(*) FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname != 'information_schema') as total_schemas,
			(SELECT COUNT(*) FROM pg_namespace WHERE nspname LIKE 'company_%') as company_schemas,
			(SELECT COUNT(*) FROM pg_namespace WHERE nspname = 'public') as public_schemas,
			(SELECT COUNT(*) FROM pg_namespace 
			 WHERE nspname NOT LIKE 'pg_%' 
			   AND nspname != 'information_schema'
			   AND nspname NOT LIKE 'company_%'
			   AND nspname != 'public') as other_schemas,
			(SELECT SUM(reltuples::bigint) FROM pg_class WHERE relkind = 'r') as total_tables,
			(SELECT COUNT(*) FROM pg_index) as total_indexes,
			(SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active') as active_conns,
			COALESCE((SELECT version FROM public.schema_migrations ORDER BY version DESC LIMIT 1), 0) as migration_version,
			COALESCE((SELECT dirty FROM public.schema_migrations ORDER BY version DESC LIMIT 1), false) as migration_dirty
	`

	err := s.pool.QueryRow(ctx, query).Scan(
		&stats.TotalDatabaseSizeMB,
		&stats.TotalSchemasCount,
		&stats.CompanySchemasCount,
		&stats.PublicSchemasCount,
		&stats.OtherSchemasCount,
		&stats.TotalTablesCount,
		&stats.TotalIndexesCount,
		&stats.TotalActiveConnections,
		&stats.MigrationVersion,
		&stats.MigrationDirty,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get system stats: %w", err)
	}

	return &stats, nil
}

func (s *Service) GetSchemaStats(ctx context.Context, schemaName string) (*SchemaStats, error) {
	var stats SchemaStats

	query := `
		SELECT 
			$1 as schema_name,
			COALESCE(
				(SELECT CAST(SUM(pg_total_relation_size(quote_ident($1) || '.' || quote_ident(tablename))) / 1024 / 1024 AS BIGINT)
				 FROM pg_tables 
				 WHERE schemaname = $1), 0
			) as schema_size_mb,
			(SELECT COUNT(*) FROM pg_tables WHERE schemaname = $1) as tables_count,
			(SELECT COUNT(*) FROM pg_indexes WHERE schemaname = $1) as indexes_count
	`

	err := s.pool.QueryRow(ctx, query, schemaName).Scan(
		&stats.SchemaName,
		&stats.SchemaSizeMB,
		&stats.TablesCount,
		&stats.IndexesCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get schema stats for %s: %w", schemaName, err)
	}

	// migration info
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM pg_tables WHERE schemaname = $1 AND tablename = 'schema_migrations')`
	err = s.pool.QueryRow(ctx, checkQuery, schemaName).Scan(&exists)
	if err == nil && exists {
		migQuery := fmt.Sprintf("SELECT version, dirty FROM %s.schema_migrations ORDER BY version DESC LIMIT 1", schemaName)
		err = s.pool.QueryRow(ctx, migQuery).Scan(&stats.MigrationVersion, &stats.MigrationDirty)
		if err != nil {
			fmt.Printf("ERROR migrating query: %v, query: %s\n", err, migQuery) // <- добавить
			stats.MigrationVersion = 0
			stats.MigrationDirty = false
		}
	}

	return &stats, nil
}

func (s *Service) GetSchemaTables(ctx context.Context, schemaName string) ([]TableInfo, error) {
	query := `
		SELECT 
			tablename as table_name,
			COALESCE(CAST(pg_total_relation_size(quote_ident($1) || '.' || quote_ident(tablename)) / 1024 AS BIGINT), 0) as size_kb,
			COALESCE(CAST(pg_total_relation_size(quote_ident($1) || '.' || quote_ident(tablename)) / 1024 / 1024 AS BIGINT), 0) as size_mb
		FROM pg_tables
		WHERE schemaname = $1
		ORDER BY size_kb DESC
	`

	rows, err := s.pool.Query(ctx, query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for schema %s: %w", schemaName, err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		err := rows.Scan(&table.TableName, &table.SizeKB, &table.SizeMB)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// helper
func quoteIdent(s string) string {
	return `"` + s + `"`
}
