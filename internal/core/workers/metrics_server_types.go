package coreworkers

import "time"

type MetricsServerSnapshot struct {
	RecordedAt time.Time `json:"recorded_at"`

	// HTTP трафик
	RequestsTotal     int `json:"requests_total"`
	Requests5xxTotal  int `json:"requests_5xx_total"`
	Requests4xxTotal  int `json:"requests_4xx_total"`
	AvgResponseTimeMs int `json:"avg_response_time_ms"`
	P95ResponseTimeMs int `json:"p95_response_time_ms"`
	ActiveConnections int `json:"active_connections"`

	// Go runtime
	GoroutinesCount int `json:"goroutines_count"`
	HeapAllocMB     int `json:"heap_alloc_mb"`
	HeapInuseMB     int `json:"heap_inuse_mb"`
	GCDurationMs    int `json:"gc_duration_ms"`

	// Воркеры
	DbWorkerSuccess        bool `json:"db_worker_success"`
	ClienteleWorkerSuccess bool `json:"clientele_worker_success"`

	// Системные
	CPUUsagePercent float64 `json:"cpu_usage_percent"`
	MemoryUsageMB   int     `json:"memory_usage_mb"`
	OpenFDsCount    int     `json:"open_fds_count"`
}
