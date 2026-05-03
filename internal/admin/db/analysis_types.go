package admindb

type SystemStats struct {
	TotalDatabaseSizeMB    int64 `json:"total_database_size_mb"`
	TotalSchemasCount      int   `json:"total_schemas_count"`
	CompanySchemasCount    int   `json:"company_schemas_count"`
	PublicSchemasCount     int   `json:"public_schemas_count"`
	OtherSchemasCount      int   `json:"other_schemas_count"`
	TotalTablesCount       int   `json:"total_tables_count"`
	TotalIndexesCount      int   `json:"total_indexes_count"`
	TotalActiveConnections int   `json:"total_active_connections"`
	MigrationVersion       int64 `json:"migration_version"`
	MigrationDirty         bool  `json:"migration_dirty"`
}

type SchemaStats struct {
	SchemaName       string `json:"schema_name"`
	SchemaSizeMB     int64  `json:"schema_size_mb"`
	TablesCount      int    `json:"tables_count"`
	IndexesCount     int    `json:"indexes_count"`
	MigrationVersion int64  `json:"migration_version"`
	MigrationDirty   bool   `json:"migration_dirty"`
}

type TableInfo struct {
	TableName string `json:"table_name"`
	SizeKB    int64  `json:"size_kb"`
	SizeMB    int64  `json:"size_mb"`
}
