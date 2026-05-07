package metrics

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "path"},
	)
)

// -----------
// SYS METRICS
// -----------

var (
	totalRequests     int64
	total5xxRequests  int64
	total4xxRequests  int64
	activeConnections int64
)

// для статуса воркеров
var (
	dbWorkerLastSuccess        int32 = 1
	clienteleWorkerLastSuccess int32 = 1
)

func SetDbWorkerLastSuccess(success bool) {
	if success {
		atomic.StoreInt32(&dbWorkerLastSuccess, 1)
	} else {
		atomic.StoreInt32(&dbWorkerLastSuccess, 0)
	}
}

func GetDbWorkerLastSuccess() bool {
	return atomic.LoadInt32(&dbWorkerLastSuccess) == 1
}

func SetClienteleWorkerLastSuccess(success bool) {
	if success {
		atomic.StoreInt32(&clienteleWorkerLastSuccess, 1)
	} else {
		atomic.StoreInt32(&clienteleWorkerLastSuccess, 0)
	}
}

func GetClienteleWorkerLastSuccess() bool {
	return atomic.LoadInt32(&clienteleWorkerLastSuccess) == 1
}

var (
	requestDurations     []float64
	requestDurationsLock sync.RWMutex
)

func GetTotalRequests() int64     { return atomic.LoadInt64(&totalRequests) }
func GetTotal5xxRequests() int64  { return atomic.LoadInt64(&total5xxRequests) }
func GetTotal4xxRequests() int64  { return atomic.LoadInt64(&total4xxRequests) }
func GetActiveConnections() int64 { return atomic.LoadInt64(&activeConnections) }

func recordRequestDuration(duration float64) {
	requestDurationsLock.Lock()
	defer requestDurationsLock.Unlock()

	requestDurations = append(requestDurations, duration)

	// Храним последние 1000 значений для p95
	if len(requestDurations) > 1000 {
		requestDurations = requestDurations[len(requestDurations)-1000:]
	}
}

func GetAvgResponseTime() int64 {
	requestDurationsLock.RLock()
	defer requestDurationsLock.RUnlock()

	if len(requestDurations) == 0 {
		return 0
	}

	var sum float64
	for _, d := range requestDurations {
		sum += d
	}

	avg := sum / float64(len(requestDurations))
	return int64(avg * 1000)
}

func GetP95ResponseTime() int64 {
	requestDurationsLock.RLock()
	defer requestDurationsLock.RUnlock()

	if len(requestDurations) == 0 {
		return 0
	}

	sorted := make([]float64, len(requestDurations))
	copy(sorted, requestDurations)
	sort.Float64s(sorted)

	idx := int(float64(len(sorted)) * 0.95)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}

	return int64(sorted[idx] * 1000)
}
