package metrics

import (
	"net"
	"net/http"
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
		// Разрешенные подсети
		allowedNetworks := []string{
			"127.0.0.1/32",   // localhost
			"::1/128",        // localhost IPv6
			"10.0.0.0/8",     // Docker сеть
			"172.16.0.0/12",  // альтернативная Docker сеть
			"192.168.0.0/16", // локальная сеть
		}

		// Получаем IP клиента
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		clientIP := net.ParseIP(host)

		// Проверяем, попадает ли IP в разрешенные подсети
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

// MetricsMiddleware возвращает функцию-мидлвар для go-chi
func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			HttpRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				http.StatusText(rw.statusCode),
			).Inc()

			HttpRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration.Seconds())
		})
	}
}
