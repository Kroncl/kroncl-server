package corestatus

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	coreworkers "kroncl-server/internal/core/workers"
	"time"
)

// GetSystemStatus возвращает текущий статус системы и историю за N дней
func (s *Service) GetSystemStatus(ctx context.Context, days int) (*SystemStatusResponse, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// получаем историю метрик сервера и БД
	serverMetrics, err := s.coreWorkers.GetServerMetricsHistory(ctx, &startDate, &endDate, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get server metrics: %w", err)
	}

	dbMetrics, err := s.coreWorkers.GetMetricsHistory(ctx, &startDate, &endDate, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get db metrics: %w", err)
	}

	// определяем инциденты
	incidents := s.detectIncidents(serverMetrics, dbMetrics)

	// группируем инциденты по дням
	incidentsByDay := make(map[string][]Incident)
	for _, inc := range incidents {
		day := inc.StartTime.Format("2006-01-02")
		incidentsByDay[day] = append(incidentsByDay[day], inc)
	}

	// строим статусы по дням
	var dailyStatuses []DailyStatus
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dayStr := currentDate.Format("2006-01-02")
		dayIncidents := incidentsByDay[dayStr]

		status := s.calculateDayStatus(dayIncidents)
		dailyStatuses = append(dailyStatuses, DailyStatus{
			Date:      dayStr,
			Status:    status,
			Incidents: dayIncidents,
		})
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// текущий статус (по последнему дню)
	currentStatus := StatusOperational
	if len(dailyStatuses) > 0 {
		currentStatus = dailyStatuses[len(dailyStatuses)-1].Status
	}

	// активные инциденты (без end_time)
	var activeIncidents []Incident
	for _, inc := range incidents {
		if inc.EndTime == nil {
			activeIncidents = append(activeIncidents, inc)
		}
	}

	return &SystemStatusResponse{
		CurrentStatus:   currentStatus,
		Dayly:           dailyStatuses,
		ActiveIncidents: activeIncidents,
	}, nil
}

// detectIncidents ищет аномалии в метриках на основе лимитов из config
func (s *Service) detectIncidents(serverMetrics []coreworkers.MetricsServerSnapshot, dbMetrics []coreworkers.MetricsDBSnapshot) []Incident {
	var incidents []Incident

	// Server метрики
	for _, m := range serverMetrics {
		// проверяем p95 response time
		if m.P95ResponseTimeMs > config.STATUS_P95_THRESHOLD_MS {
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-p95-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    s.getSeverityByP95(m.P95ResponseTimeMs),
				Title:       "Высокое время ответа API",
				Description: fmt.Sprintf("P95 время ответа достигло %d мс (порог: %d мс)", m.P95ResponseTimeMs, config.STATUS_P95_THRESHOLD_MS),
				MetricsType: "server",
			})
		}

		// проверяем 5xx ошибки
		if m.Requests5xxTotal > config.STATUS_5XX_THRESHOLD_PER_MINUTE {
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-5xx-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    SeverityMajor,
				Title:       "Много ошибок 5xx",
				Description: fmt.Sprintf("За минуту зафиксировано %d ошибок 5xx (порог: %d)", m.Requests5xxTotal, config.STATUS_5XX_THRESHOLD_PER_MINUTE),
				MetricsType: "server",
			})
		}

		// проверяем GC паузы
		if m.GCDurationMs > config.STATUS_GC_THRESHOLD_MS {
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-gc-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    s.getSeverityByGC(m.GCDurationMs),
				Title:       "Длительная GC пауза",
				Description: fmt.Sprintf("GC пауза составила %d мс (порог: %d мс)", m.GCDurationMs, config.STATUS_GC_THRESHOLD_MS),
				MetricsType: "server",
			})
		}
	}

	// DB метрики (упрощённо)
	for _, m := range dbMetrics {
		// проверяем активные соединения
		if m.TotalActiveConnections > config.STATUS_DB_CONNECTIONS_THRESHOLD {
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("db-conn-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    SeverityMajor,
				Title:       "Много активных соединений с БД",
				Description: fmt.Sprintf("Активных соединений: %d (порог: %d)", m.TotalActiveConnections, config.STATUS_DB_CONNECTIONS_THRESHOLD),
				MetricsType: "db",
			})
		}
	}

	return s.mergeAdjacentIncidents(incidents, 5*time.Minute)
}

// mergeAdjacentIncidents объединяет инциденты, которые произошли в течение window друг от друга
func (s *Service) mergeAdjacentIncidents(incidents []Incident, window time.Duration) []Incident {
	if len(incidents) == 0 {
		return incidents
	}

	// сортируем по времени
	sorted := make([]Incident, len(incidents))
	copy(sorted, incidents)

	var merged []Incident
	current := sorted[0]

	for i := 1; i < len(sorted); i++ {
		next := sorted[i]
		if next.StartTime.Sub(current.StartTime) <= window && next.Title == current.Title {
			current.EndTime = &next.StartTime
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)

	return merged
}

func (s *Service) calculateDayStatus(incidents []Incident) Status {
	if len(incidents) == 0 {
		return StatusOperational
	}

	for _, inc := range incidents {
		if inc.Severity == SeverityMajor {
			return StatusPartialOutage
		}
	}
	return StatusDegraded
}

func (s *Service) getSeverityByP95(p95Ms int) IncidentSeverity {
	if p95Ms > config.STATUS_P95_CRITICAL_MS {
		return SeverityMajor
	}
	return SeverityMinor
}

func (s *Service) getSeverityByGC(gcMs int) IncidentSeverity {
	if gcMs > config.STATUS_GC_CRITICAL_MS {
		return SeverityMajor
	}
	return SeverityMinor
}
