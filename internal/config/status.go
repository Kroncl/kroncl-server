package config

const (
	// server
	STATUS_P95_THRESHOLD_MS         = 500  // ms, превышение -> инцидент
	STATUS_P95_CRITICAL_MS          = 1000 // ms, критический инцидент
	STATUS_5XX_THRESHOLD_PER_MINUTE = 10   // штук, превышение -> инцидент
	STATUS_GC_THRESHOLD_MS          = 100  // ms, превышение -> инцидент
	STATUS_GC_CRITICAL_MS           = 300  // ms, критический инцидент

	// db
	STATUS_DB_CONNECTIONS_THRESHOLD = 100 // активных соединений
)
