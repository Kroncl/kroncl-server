package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func NewClient(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = "https://goapi.unisender.ru/ru/transactional/api/v1/email"
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// Do выполняет запрос к API
func (c *Client) Do(ctx context.Context, method string, reqBody interface{}, respBody interface{}) error {
	// Формируем URL
	url := fmt.Sprintf("%s%s.json", c.baseURL, method)

	// Маршалим тело
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)

	// Отправляем
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Проверяем статус
	if err := c.handleStatusCode(resp.StatusCode, body, url); err != nil {
		return err
	}

	// Парсим ответ
	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) handleStatusCode(statusCode int, body []byte, url string) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return c.parseError(body, "bad request")
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized: invalid API key")
	case http.StatusForbidden:
		return fmt.Errorf("forbidden: no permissions")
	case http.StatusNotFound:
		return fmt.Errorf("not found: invalid method: %s", url)
	case http.StatusRequestEntityTooLarge:
		return fmt.Errorf("request too large: max 10MB")
	case http.StatusTooManyRequests:
		return fmt.Errorf("too many requests: rate limit exceeded")
	default:
		if statusCode >= 500 && statusCode < 600 {
			return fmt.Errorf("server error (HTTP %d): retry later", statusCode)
		}
		return c.parseError(body, fmt.Sprintf("unexpected status code %d", statusCode))
	}
}

func (c *Client) parseError(body []byte, defaultMsg string) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
		return NewAPIError(errResp.Code, errResp.Message)
	}
	return fmt.Errorf("%s: %s", defaultMsg, string(body))
}
