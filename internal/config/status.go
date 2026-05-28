package config

const (
	// server-limits
	STATUS_P95_THRESHOLD_MS = 500  // ms, превышение -> minor incident
	STATUS_P95_CRITICAL_MS  = 1000 // ms, превышение -> major incident

	STATUS_5XX_THRESHOLD_PER_MINUTE = 10 // штук, превышение -> minor incident
	STATUS_5XX_CRITICAL_PER_MINUTE  = 50 // штук, превышение -> major incident

	STATUS_GC_THRESHOLD_MS = 100 // ms, превышение -> minor incident
	STATUS_GC_CRITICAL_MS  = 300 // ms, превышение -> major incident

	STATUS_AVG_RESPONSE_THRESHOLD_MS = 200 // ms, превышение -> minor incident
	STATUS_AVG_RESPONSE_CRITICAL_MS  = 500 // ms, превышение -> major incident

	STATUS_CPU_THRESHOLD_PERCENT = 70.0 // %, превышение -> minor incident
	STATUS_CPU_CRITICAL_PERCENT  = 90.0 // %, превышение -> major incident

	STATUS_MEMORY_THRESHOLD_MB = 1024 // MB, превышение -> minor incident
	STATUS_MEMORY_CRITICAL_MB  = 2048 // MB, превышение -> major incident

	STATUS_GOROUTINES_THRESHOLD = 5000  // штук, превышение -> minor incident
	STATUS_GOROUTINES_CRITICAL  = 10000 // штук, превышение -> major incident

	STATUS_OPEN_FDS_THRESHOLD = 500  // штук, превышение -> minor incident
	STATUS_OPEN_FDS_CRITICAL  = 1000 // штук, превышение -> major incident

	// db-limits
	STATUS_DB_CONNECTIONS_THRESHOLD = 50  // активных соединений, превышение -> minor
	STATUS_DB_CONNECTIONS_CRITICAL  = 100 // активных соединений, превышение -> major

	STATUS_DB_ROLLBACK_RATIO_THRESHOLD = 0.05 // 5% от всех транзакций, превышение -> minor
	STATUS_DB_ROLLBACK_RATIO_CRITICAL  = 0.15 // 15% от всех транзакций, превышение -> major

	STATUS_DB_CACHE_HIT_RATIO_THRESHOLD = 0.95 // 95%, ниже -> minor
	STATUS_DB_CACHE_HIT_RATIO_CRITICAL  = 0.90 // 90%, ниже -> major

	STATUS_DB_TUPLE_RATE_THRESHOLD = 50000  // tuples/sec, превышение -> minor
	STATUS_DB_TUPLE_RATE_CRITICAL  = 100000 // tuples/sec, превышение -> major

	// media-limits (объектное хранилище)
	STATUS_MEDIA_OBJECTS_GROWTH_THRESHOLD = 1000   // объектов за час, превышение -> minor
	STATUS_MEDIA_OBJECTS_GROWTH_CRITICAL  = 5000   // объектов за час, превышение -> major
	STATUS_MEDIA_SIZE_GROWTH_THRESHOLD    = 1024   // MB за час, превышение -> minor
	STATUS_MEDIA_SIZE_GROWTH_CRITICAL     = 5120   // MB за час, превышение -> major
	STATUS_MEDIA_TENANT_OBJECTS_THRESHOLD = 50000  // объектов на тенант, превышение -> minor
	STATUS_MEDIA_TENANT_OBJECTS_CRITICAL  = 200000 // объектов на тенант, превышение -> major
	STATUS_MEDIA_TENANT_SIZE_THRESHOLD    = 10240  // MB на тенант, превышение -> minor
	STATUS_MEDIA_TENANT_SIZE_CRITICAL     = 51200  // MB на тенант, превышение -> major
)
