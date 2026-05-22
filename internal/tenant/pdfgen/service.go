package pdfgen

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/nativebpm/gotenberg-client"
)

func NewService(cfg Config) (*Service, error) {
	client, err := gotenberg.NewClient(http.Client{}, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create gotenberg client: %w", err)
	}

	if cfg.TemplatesPath == "" {
		cfg.TemplatesPath = "./templates"
	}

	if _, err := os.Stat(cfg.TemplatesPath); err != nil {
		return nil, fmt.Errorf("templates directory not found: %w", err)
	}

	return &Service{
		client:      client,
		config:      cfg,
		templateDir: cfg.TemplatesPath,
	}, nil
}

func (s *Service) GenerateFromTemplate(ctx context.Context, req GenerateFromTemplateRequest) (*bytes.Buffer, error) {
	fullPath := filepath.Join(s.templateDir, req.TemplatePath)

	// Парсим шаблон с функциями
	tmpl, err := template.New(filepath.Base(req.TemplatePath)).Funcs(getFuncMap()).ParseFiles(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, req.Data); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Добавляем таймаут для запроса
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	chromium := s.client.Chromium()

	// Явно указываем convert HTML to PDF
	convertReq := chromium.ConvertHTML(ctx, &htmlBuf)

	// Устанавливаем таймаут
	convertReq.Timeout(60 * time.Second)

	// Устанавливаем параметры
	if req.Options != nil {
		switch req.Options.PaperSize {
		case "A4":
			convertReq.PaperSizeA4()
		case "Letter":
			convertReq.PaperSizeLetter()
		default:
			convertReq.PaperSizeA4()
		}

		if req.Options.Landscape {
			convertReq.Landscape()
		}

		convertReq.Margins(
			req.Options.MarginTop,
			req.Options.MarginBottom,
			req.Options.MarginLeft,
			req.Options.MarginRight,
		)
	} else {
		convertReq.PaperSizeA4().Margins(0.5, 0.5, 0.5, 0.5)
	}

	// Добавляем заголовок для PDF
	convertReq.Header("Content-Type", "application/pdf")
	convertReq.OutputFilename("invoice.pdf")

	log.Printf("HTML to convert (first 500 chars): %s", htmlBuf.String()[:min(500, htmlBuf.Len())])

	// Проверь, что Gotenberg доступен
	if _, err := http.Get(s.config.Endpoint + "/health"); err != nil {
		return nil, fmt.Errorf("Gotenberg not reachable: %w", err)
	}

	resp, err := convertReq.Send()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем Content-Type ответа
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		// Читаем немного тела для отладки
		bodyPreview := make([]byte, 200)
		resp.Body.Read(bodyPreview)
		return nil, fmt.Errorf("expected PDF but got %s: %s", contentType, string(bodyPreview))
	}

	var pdfBuf bytes.Buffer
	if _, err := pdfBuf.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read PDF response: %w", err)
	}

	// Проверяем, что это действительно PDF файл (магическое число %PDF)
	if pdfBuf.Len() < 4 || string(pdfBuf.Bytes()[:4]) != "%PDF" {
		return nil, fmt.Errorf("generated file is not a valid PDF")
	}

	return &pdfBuf, nil
}
