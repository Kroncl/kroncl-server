package corestatus

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	coreworkers "kroncl-server/internal/core/workers"
	"time"
)

func (s *Service) GetComponentStatus(ctx context.Context, compType ComponentType, days int) ([]DailyStatus, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// всегда тянем полные данные
	serverMetrics, err := s.coreWorkers.GetServerMetricsHistory(ctx, &startDate, &endDate, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get server metrics: %w", err)
	}

	dbMetrics, err := s.coreWorkers.GetMetricsHistory(ctx, &startDate, &endDate, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get db metrics: %w", err)
	}

	// определяем инциденты
	serverIncidents := s.detectServerIncidents(serverMetrics)
	dbIncidents := s.detectDBIncidents(dbMetrics)

	// группируем по дням
	serverIncidentsByDay := s.groupIncidentsByDay(serverIncidents)
	dbIncidentsByDay := s.groupIncidentsByDay(dbIncidents)

	// строим статусы по дням
	var dailyStatuses []DailyStatus
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dayStr := currentDate.Format("2006-01-02")

		var incidents []Incident
		var status Status = StatusOperational

		switch compType {
		case ComponentAll:
			incidents = append(serverIncidentsByDay[dayStr], dbIncidentsByDay[dayStr]...)
			status = s.calculateOverallStatus(serverIncidentsByDay[dayStr], dbIncidentsByDay[dayStr])
		case ComponentServer:
			incidents = serverIncidentsByDay[dayStr]
			status = s.calculateStatusFromIncidents(incidents)
		case ComponentStorage:
			incidents = dbIncidentsByDay[dayStr]
			status = s.calculateStatusFromIncidents(incidents)
		}

		dailyStatuses = append(dailyStatuses, DailyStatus{
			Date:      dayStr,
			Status:    status,
			Incidents: incidents,
		})
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return dailyStatuses, nil
}

func (s *Service) GetFullSystemStatus(ctx context.Context, days int) (*SystemStatusResponse, error) {
	allDaily, err := s.GetComponentStatus(ctx, ComponentAll, days)
	if err != nil {
		return nil, err
	}

	serverDaily, err := s.GetComponentStatus(ctx, ComponentServer, days)
	if err != nil {
		return nil, err
	}

	storageDaily, err := s.GetComponentStatus(ctx, ComponentStorage, days)
	if err != nil {
		return nil, err
	}

	// текущий статус (по последнему дню)
	currentStatus := StatusOperational
	if len(allDaily) > 0 {
		currentStatus = allDaily[len(allDaily)-1].Status
	}

	// активные инциденты
	var activeIncidents []Incident
	for _, day := range allDaily {
		for _, inc := range day.Incidents {
			if inc.EndTime == nil {
				activeIncidents = append(activeIncidents, inc)
			}
		}
	}

	return &SystemStatusResponse{
		CurrentStatus:   currentStatus,
		Daily:           allDaily,
		ActiveIncidents: activeIncidents,
		Components: map[ComponentType][]DailyStatus{
			ComponentAll:     allDaily,
			ComponentServer:  serverDaily,
			ComponentStorage: storageDaily,
		},
	}, nil
}
func (s *Service) groupIncidentsByDay(incidents []Incident) map[string][]Incident {
	result := make(map[string][]Incident)
	for _, inc := range incidents {
		day := inc.StartTime.Format("2006-01-02")
		result[day] = append(result[day], inc)
	}
	return result
}

func (s *Service) calculateOverallStatus(serverIncidents, dbIncidents []Incident) Status {
	allIncidents := append(serverIncidents, dbIncidents...)
	return s.calculateStatusFromIncidents(allIncidents)
}

func (s *Service) calculateStatusFromIncidents(incidents []Incident) Status {
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

func (s *Service) mergeAdjacentIncidents(incidents []Incident, window time.Duration) []Incident {
	if len(incidents) == 0 {
		return incidents
	}

	// сортируем по времени
	sorted := make([]Incident, len(incidents))
	copy(sorted, incidents)
	// для простоты предполагаем, что данные уже отсортированы по start_time

	var merged []Incident
	current := sorted[0]

	for i := 1; i < len(sorted); i++ {
		next := sorted[i]
		// если следующий инцидент того же типа и произошёл в течение window
		if next.StartTime.Sub(current.StartTime) <= window && next.Title == current.Title {
			// расширяем текущий инцидент
			current.EndTime = &next.StartTime
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)

	return merged
}

func (s *Service) detectServerIncidents(metrics []coreworkers.MetricsServerSnapshot) []Incident {
	var incidents []Incident
	for _, m := range metrics {
		// P95 response time
		if m.P95ResponseTimeMs > config.STATUS_P95_THRESHOLD_MS {
			severity := s.getSeverityByP95(m.P95ResponseTimeMs)
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-p95-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Высокое время ответа API",
				Description: fmt.Sprintf("P95 время ответа достигло %d мс", m.P95ResponseTimeMs),
				MetricsType: "server",
			})
		}

		// 5xx errors
		if m.Requests5xxTotal > config.STATUS_5XX_THRESHOLD_PER_MINUTE {
			severity := SeverityMinor
			if m.Requests5xxTotal > config.STATUS_5XX_CRITICAL_PER_MINUTE {
				severity = SeverityMajor
			}
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-5xx-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Много ошибок 5xx",
				Description: fmt.Sprintf("За минуту зафиксировано %d ошибок 5xx", m.Requests5xxTotal),
				MetricsType: "server",
			})
		}

		// GC duration
		if m.GCDurationMs > config.STATUS_GC_THRESHOLD_MS {
			severity := s.getSeverityByGC(m.GCDurationMs)
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-gc-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Длительная GC пауза",
				Description: fmt.Sprintf("GC пауза составила %d мс", m.GCDurationMs),
				MetricsType: "server",
			})
		}

		// Avg response time
		if m.AvgResponseTimeMs > config.STATUS_AVG_RESPONSE_THRESHOLD_MS {
			severity := SeverityMinor
			if m.AvgResponseTimeMs > config.STATUS_AVG_RESPONSE_CRITICAL_MS {
				severity = SeverityMajor
			}
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-avg-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Высокое среднее время ответа",
				Description: fmt.Sprintf("Среднее время ответа: %d мс", m.AvgResponseTimeMs),
				MetricsType: "server",
			})
		}

		// CPU usage
		if m.CPUUsagePercent > config.STATUS_CPU_THRESHOLD_PERCENT {
			severity := SeverityMinor
			if m.CPUUsagePercent > config.STATUS_CPU_CRITICAL_PERCENT {
				severity = SeverityMajor
			}
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-cpu-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Высокая загрузка CPU",
				Description: fmt.Sprintf("Загрузка CPU: %.1f%%", m.CPUUsagePercent),
				MetricsType: "server",
			})
		}

		// Memory usage
		if m.MemoryUsageMB > config.STATUS_MEMORY_THRESHOLD_MB {
			severity := SeverityMinor
			if m.MemoryUsageMB > config.STATUS_MEMORY_CRITICAL_MB {
				severity = SeverityMajor
			}
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-mem-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Высокое потребление памяти",
				Description: fmt.Sprintf("RSS память: %d MB", m.MemoryUsageMB),
				MetricsType: "server",
			})
		}

		// Goroutines leak
		if m.GoroutinesCount > config.STATUS_GOROUTINES_THRESHOLD {
			severity := SeverityMinor
			if m.GoroutinesCount > config.STATUS_GOROUTINES_CRITICAL {
				severity = SeverityMajor
			}
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-goroutines-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Утечка горутин",
				Description: fmt.Sprintf("Количество горутин: %d", m.GoroutinesCount),
				MetricsType: "server",
			})
		}

		// Open file descriptors
		if m.OpenFDsCount > config.STATUS_OPEN_FDS_THRESHOLD {
			severity := SeverityMinor
			if m.OpenFDsCount > config.STATUS_OPEN_FDS_CRITICAL {
				severity = SeverityMajor
			}
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("server-fds-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Много открытых файловых дескрипторов",
				Description: fmt.Sprintf("Открыто FD: %d", m.OpenFDsCount),
				MetricsType: "server",
			})
		}
	}
	return s.mergeAdjacentIncidents(incidents, 5*time.Minute)
}

func (s *Service) detectDBIncidents(metrics []coreworkers.MetricsDBSnapshot) []Incident {
	var incidents []Incident
	for _, m := range metrics {
		// Active connections
		if m.TotalActiveConnections > config.STATUS_DB_CONNECTIONS_THRESHOLD {
			severity := SeverityMinor
			if m.TotalActiveConnections > config.STATUS_DB_CONNECTIONS_CRITICAL {
				severity = SeverityMajor
			}
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("db-conn-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Много активных соединений с БД",
				Description: fmt.Sprintf("Активных соединений: %d", m.TotalActiveConnections),
				MetricsType: "db",
			})
		}

		// Rollback ratio
		totalTx := m.XactCommit + m.XactRollback
		if totalTx > 0 {
			rollbackRatio := float64(m.XactRollback) / float64(totalTx)
			if rollbackRatio > config.STATUS_DB_ROLLBACK_RATIO_THRESHOLD {
				severity := SeverityMinor
				if rollbackRatio > config.STATUS_DB_ROLLBACK_RATIO_CRITICAL {
					severity = SeverityMajor
				}
				incidents = append(incidents, Incident{
					ID:          fmt.Sprintf("db-rollback-%d", m.RecordedAt.Unix()),
					StartTime:   m.RecordedAt,
					Severity:    severity,
					Title:       "Высокий процент откатов транзакций",
					Description: fmt.Sprintf("Доля откатов: %.1f%%", rollbackRatio*100),
					MetricsType: "db",
				})
			}
		}

		// Cache hit ratio
		totalIO := m.BlksHit + m.BlksRead
		if totalIO > 0 {
			cacheHitRatio := float64(m.BlksHit) / float64(totalIO)
			if cacheHitRatio < config.STATUS_DB_CACHE_HIT_RATIO_THRESHOLD {
				severity := SeverityMinor
				if cacheHitRatio < config.STATUS_DB_CACHE_HIT_RATIO_CRITICAL {
					severity = SeverityMajor
				}
				incidents = append(incidents, Incident{
					ID:          fmt.Sprintf("db-cache-%d", m.RecordedAt.Unix()),
					StartTime:   m.RecordedAt,
					Severity:    severity,
					Title:       "Низкая эффективность кэша БД",
					Description: fmt.Sprintf("Cache hit ratio: %.1f%%", cacheHitRatio*100),
					MetricsType: "db",
				})
			}
		}

		// High tuple modification rate (possible write spike)
		tupleRate := m.TupInserted + m.TupUpdated + m.TupDeleted
		if tupleRate > config.STATUS_DB_TUPLE_RATE_THRESHOLD {
			severity := SeverityMinor
			if tupleRate > config.STATUS_DB_TUPLE_RATE_CRITICAL {
				severity = SeverityMajor
			}
			incidents = append(incidents, Incident{
				ID:          fmt.Sprintf("db-tuple-rate-%d", m.RecordedAt.Unix()),
				StartTime:   m.RecordedAt,
				Severity:    severity,
				Title:       "Высокая интенсивность изменений данных",
				Description: fmt.Sprintf("Изменено строк: %d", tupleRate),
				MetricsType: "db",
			})
		}
	}
	return s.mergeAdjacentIncidents(incidents, 5*time.Minute)
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
