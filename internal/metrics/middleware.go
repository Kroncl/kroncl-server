package metrics

import (
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func PrometheusIPWhitelist(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedNetworks := []string{
			"127.0.0.1/32",   // localhost
			"::1/128",        // localhost IPv6
			"10.0.0.0/8",     // Docker сеть
			"172.16.0.0/12",  // альтернативная Docker сеть
			"192.168.0.0/16", // локальная сеть
		}

		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		clientIP := net.ParseIP(host)

		allowed := false
		for _, cidr := range allowedNetworks {
			_, ipnet, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			if ipnet.Contains(clientIP) {
				allowed = true
				break
			}
		}

		if !allowed {
			http.Error(w, "Forbidden: access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&activeConnections, 1)
			defer atomic.AddInt64(&activeConnections, -1)

			atomic.AddInt64(&totalRequests, 1)

			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Prometheus метрики (уже зарегистрированы через promauto в metrics.go)
			HttpRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				http.StatusText(rw.statusCode),
			).Inc()

			HttpRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration.Seconds())

			// Записываем длительность для p95
			recordRequestDuration(duration.Seconds())

			// Считаем 4xx и 5xx
			if rw.statusCode >= 500 {
				atomic.AddInt64(&total5xxRequests, 1)
			} else if rw.statusCode >= 400 {
				atomic.AddInt64(&total4xxRequests, 1)
			}
		})
	}
}
