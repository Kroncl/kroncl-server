package coreworkers

import "time"

type MetricsDBSnapshot struct {
	RecordedAt             time.Time `json:"recorded_at"`
	TotalDatabaseSizeMB    int64     `json:"total_database_size_mb"`
	TotalSchemasCount      int       `json:"total_schemas_count"`
	CompanySchemasCount    int       `json:"company_schemas_count"`
	TotalTablesCount       int       `json:"total_tables_count"`
	TotalIndexesCount      int       `json:"total_indexes_count"`
	TotalActiveConnections int       `json:"total_active_connections"`
	XactCommit             int64     `json:"xact_commit"`
	XactRollback           int64     `json:"xact_rollback"`
	TupReturned            int64     `json:"tup_returned"`
	TupFetched             int64     `json:"tup_fetched"`
	TupInserted            int64     `json:"tup_inserted"`
	TupUpdated             int64     `json:"tup_updated"`
	TupDeleted             int64     `json:"tup_deleted"`
	BlksRead               int64     `json:"blks_read"`
	BlksHit                int64     `json:"blks_hit"`
}
