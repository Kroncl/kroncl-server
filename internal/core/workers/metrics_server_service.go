package coreworkers

import (
	"context"
	"fmt"
	"kroncl-server/internal/metrics"
	"runtime"
	"time"
)

func (s *Service) CollectServerMetrics(ctx context.Context) (*MetricsServerSnapshot, error) {
	stats := &MetricsServerSnapshot{
		RecordedAt: time.Now(),
	}

	// дельты (прирост за интервал)
	stats.RequestsTotal = int(metrics.GetRequestsDelta())
	stats.Requests5xxTotal = int(metrics.Get5xxDelta())
	stats.Requests4xxTotal = int(metrics.Get4xxDelta())

	// gauges (текущие значения)
	stats.ActiveConnections = int(metrics.GetActiveConnections())
	stats.AvgResponseTimeMs = int(metrics.GetAvgResponseTime())
	stats.P95ResponseTimeMs = int(metrics.GetP95ResponseTime())

	// Runtime метрики
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	stats.HeapAllocMB = int(memStats.HeapAlloc / 1024 / 1024)
	stats.HeapInuseMB = int(memStats.HeapInuse / 1024 / 1024)
	stats.GoroutinesCount = runtime.NumGoroutine()
	stats.GCDurationMs = int(memStats.PauseTotalNs / 1e6)

	// Статус воркеров
	stats.DbWorkerSuccess = metrics.GetDbWorkerLastSuccess()
	stats.ClienteleWorkerSuccess = metrics.GetClienteleWorkerLastSuccess()

	// Системные метрики
	stats.OpenFDsCount = getOpenFDs()
	stats.MemoryUsageMB = getMemoryUsage()
	stats.CPUUsagePercent = getCPUUsage()

	return stats, nil
}

func (s *Service) SaveServerMetricsSnapshot(ctx context.Context, stats *MetricsServerSnapshot) error {
	query := `
        INSERT INTO metrics_server_history (
            recorded_at,
            requests_total, requests_5xx_total, requests_4xx_total,
            avg_response_time_ms, p95_response_time_ms, active_connections,
            goroutines_count, heap_alloc_mb, heap_inuse_mb, gc_duration_ms,
            db_worker_success, clientele_worker_success,
            cpu_usage_percent, memory_usage_mb, open_fds_count
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
    `

	_, err := s.pool.Exec(ctx, query,
		stats.RecordedAt,
		stats.RequestsTotal, stats.Requests5xxTotal, stats.Requests4xxTotal,
		stats.AvgResponseTimeMs, stats.P95ResponseTimeMs, stats.ActiveConnections,
		stats.GoroutinesCount, stats.HeapAllocMB, stats.HeapInuseMB, stats.GCDurationMs,
		stats.DbWorkerSuccess, stats.ClienteleWorkerSuccess,
		stats.CPUUsagePercent, stats.MemoryUsageMB, stats.OpenFDsCount,
	)

	if err != nil {
		return fmt.Errorf("failed to save server metrics snapshot: %w", err)
	}

	return nil
}

// GetServerMetricsHistory возвращает историю метрик сервера с фильтрацией
func (s *Service) GetServerMetricsHistory(ctx context.Context, startDate, endDate *time.Time, limit int) ([]MetricsServerSnapshot, error) {
	query := `
        SELECT 
            recorded_at,
            requests_total, requests_5xx_total, requests_4xx_total,
            avg_response_time_ms, p95_response_time_ms, active_connections,
            goroutines_count, heap_alloc_mb, heap_inuse_mb, gc_duration_ms,
            db_worker_success, clientele_worker_success,
            cpu_usage_percent, memory_usage_mb, open_fds_count
        FROM metrics_server_history
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
		return nil, fmt.Errorf("failed to get server metrics history: %w", err)
	}
	defer rows.Close()

	var serverMetrics []MetricsServerSnapshot
	for rows.Next() {
		var m MetricsServerSnapshot
		err := rows.Scan(
			&m.RecordedAt,
			&m.RequestsTotal, &m.Requests5xxTotal, &m.Requests4xxTotal,
			&m.AvgResponseTimeMs, &m.P95ResponseTimeMs, &m.ActiveConnections,
			&m.GoroutinesCount, &m.HeapAllocMB, &m.HeapInuseMB, &m.GCDurationMs,
			&m.DbWorkerSuccess, &m.ClienteleWorkerSuccess,
			&m.CPUUsagePercent, &m.MemoryUsageMB, &m.OpenFDsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server metric: %w", err)
		}
		serverMetrics = append(serverMetrics, m)
	}

	return serverMetrics, nil
}
