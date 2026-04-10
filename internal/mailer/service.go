package mailer

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"log"
)

type Service struct {
	config *config.MailSenderConfig
	client *Client
}

func NewService(cfg *config.MailSenderConfig) *Service {
	return &Service{
		config: cfg,
		client: NewClient(cfg.ApiUrl, cfg.ApiKey),
	}
}

const (
	METHOD_EMAIL_SEND = "/email/send"
)

// SendSimple отправляет простое письмо
func (s *Service) SendSimple(ctx context.Context, to, subject, htmlBody, plainText string) (*SendResponse, error) {
	msg := Message{
		Recipients: []Recipient{{Email: to}},
		Subject:    subject,
		FromEmail:  s.config.NotifyDomain,
		FromName:   "Kroncl",
		Body: Body{
			HTML:      htmlBody,
			Plaintext: plainText,
		},
		TrackLinks: 1,
		TrackRead:  1,
	}

	return s.Send(ctx, msg)
}

// Send отправляет произвольное сообщение
func (s *Service) Send(ctx context.Context, msg Message) (*SendResponse, error) {
	// Валидация
	if len(msg.Recipients) == 0 {
		return nil, fmt.Errorf("no recipients specified")
	}
	if msg.Subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	if msg.FromEmail == "" {
		msg.FromEmail = s.config.NotifyDomain
	}
	if msg.Body.HTML == "" && msg.Body.Plaintext == "" {
		return nil, fmt.Errorf("either html or plaintext body is required")
	}

	// Отправляем
	req := SendRequest{Message: msg}
	var resp SendResponse

	if err := s.client.Do(ctx, METHOD_EMAIL_SEND, req, &resp); err != nil {
		log.Printf("❌ Unisender send error: %v", err)
		return nil, err
	}

	// Логируем результат
	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Email sent with failures: job_id=%s, success=%d, failed=%d",
			resp.JobID, len(resp.Emails), len(resp.FailedEmails))
	} else {
		log.Printf("✅ Email sent: job_id=%s, recipient=%s", resp.JobID, msg.Recipients[0].Email)
	}

	return &resp, nil
}
