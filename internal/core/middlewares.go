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

// Response стандартная структура ответа API
type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
	Meta    Meta        `json:"meta"`
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
		// Пропускаем метрики без изменений
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Пропускаем health check
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Создаем кастомный writer
		crw := &capturingResponseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
		}

		// Выполняем хендлер
		next.ServeHTTP(crw, r)

		// Если статус не установлен, ставим 200
		if crw.statusCode == 0 {
			crw.statusCode = http.StatusOK
		}

		// Парсим JSON из ответа хендлера
		var data interface{}
		var message string
		rawBody := crw.body.String()

		// Если ответ пустой
		if rawBody == "" {
			if crw.statusCode < 400 {
				data = map[string]interface{}{}
			}
		} else {
			// Пробуем распарсить как JSON
			var parsedData map[string]interface{}
			if err := json.Unmarshal([]byte(rawBody), &parsedData); err == nil {
				// Если это JSON, извлекаем message если есть
				if msg, ok := parsedData["message"].(string); ok {
					message = msg
					delete(parsedData, "message") // Убираем message
				}
				if msg, ok := parsedData["error"].(string); ok && message == "" {
					message = msg
					delete(parsedData, "error") // Убираем error
				}

				// ОСНОВНОЕ ИСПРАВЛЕНИЕ:
				// Если есть поле "data" в ответе хендлера, берем его содержимое
				if handlerData, ok := parsedData["data"]; ok {
					data = handlerData
				} else if len(parsedData) > 0 {
					// Если остались другие поля после удаления message/error
					data = parsedData
				} else if crw.statusCode < 400 {
					// Для успешных ответов без данных
					data = map[string]interface{}{}
				}
			} else {
				// Если не JSON, используем как строку для сообщения
				message = strings.TrimSpace(rawBody)
				data = map[string]interface{}{}
			}
		}

		// Автоматически генерируем сообщение если его нет
		if message == "" {
			if crw.statusCode < 400 {
				// Для успешных ответов
				switch r.Method {
				case http.MethodPost:
					message = "Created successfully"
				case http.MethodPut, http.MethodPatch:
					message = "Updated successfully"
				case http.MethodDelete:
					message = "Deleted successfully"
				default:
					message = "Success"
				}
			} else {
				// Для ошибок
				switch crw.statusCode {
				case http.StatusBadRequest:
					message = "Bad request"
				case http.StatusUnauthorized:
					message = "Unauthorized"
				case http.StatusForbidden:
					message = "Forbidden"
				case http.StatusNotFound:
					message = "Not found"
				case http.StatusInternalServerError:
					message = "Internal server error"
				default:
					message = "Error"
				}
			}
		}

		// Формируем стандартный ответ
		response := Response{
			Status:  crw.statusCode < 400,
			Message: message,
			Data:    data,
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
