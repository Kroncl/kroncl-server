package core

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type Response struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
	Meta   Meta        `json:"_meta"`
}

type Meta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
	Path      string `json:"path"`
	Method    string `json:"method"`
}

// BaseResponse middleware для стандартного формата ответа
func BaseResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем health check
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Создаем кастомный writer который перехватывает вывод
		crw := &capturingResponseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
		}

		// Выполняем хендлер
		next.ServeHTTP(crw, r)

		// Если ничего не записали (пустой ответ)
		if crw.body.Len() == 0 {
			crw.body.WriteString("{}") // пустой JSON
		}

		// После получения данных:
		var data interface{}
		if err := json.Unmarshal(crw.body.Bytes(), &data); err != nil {
			// Если не JSON, используем как строку и чистим \n
			rawString := crw.body.String()
			// Убираем все \n и \r из конца строки
			rawString = strings.TrimRight(rawString, "\r\n")
			data = rawString
		}

		// Формируем стандартный ответ
		response := Response{
			Status: crw.statusCode < 400,
			Data:   data,
			Meta: Meta{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				RequestID: GetRequestID(r.Context()),
				Path:      r.URL.Path,
				Method:    r.Method,
			},
		}

		// Отправляем финальный JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(crw.statusCode)
		json.NewEncoder(w).Encode(response)
	})
}

// capturingResponseWriter перехватывает вывод
type capturingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (crw *capturingResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
}

func (crw *capturingResponseWriter) Write(b []byte) (int, error) {
	// Если статус не установлен, по умолчанию 200 OK
	if crw.statusCode == 0 {
		crw.statusCode = http.StatusOK
	}

	// Записываем в буфер
	return crw.body.Write(b)
}

func (crw *capturingResponseWriter) Header() http.Header {
	return crw.ResponseWriter.Header()
}

// GetRequestID получает ID запроса из контекста
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(middleware.RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}
