package mailer

import (
	"context"
	"fmt"
	"log"
	"time"
)

type AdminSupportMessageData struct {
	AdminEmails  []string
	Message      string
	CompanyName  string
	AccountName  string
	AccountEmail string
	TicketID     string
}

func (s *Service) SendAdminSupportNewMessage(ctx context.Context, data *AdminSupportMessageData) {
	if len(data.AdminEmails) == 0 {
		log.Printf("❌ No admin emails provided for support message")
		return
	}

	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := fmt.Sprintf("[Тикет #%s] Новое сообщение в поддержку от %s (%s)", data.TicketID, data.CompanyName, data.AccountName)

	htmlBody := fmt.Sprintf(`
		<h2>Новое сообщение в техническую поддержку</h2>
		
		<p><strong>Тикет:</strong> #%s</p>
		
		<p><strong>Отправитель:</strong><br>
		Компания: %s<br>
		Пользователь: %s<br>
		Email: <a href="mailto:%s">%s</a></p>
		
		<p><strong>Сообщение:</strong><br>
		<div style="background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin: 10px 0;">
			%s
		</div>
		</p>
		
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">Перейдите на платформу администраторов, чтобы ответить на сообщение клиента.<br>
		С уважением,<br>Kroncl</p>
	`, data.TicketID, data.CompanyName, data.AccountName, data.AccountEmail, data.AccountEmail, data.Message)

	plainTextBody := fmt.Sprintf(
		"Новое сообщение в техническую поддержку\n\n"+
			"Тикет: #%s\n\n"+
			"Отправитель:\n"+
			"Компания: %s\n"+
			"Пользователь: %s\n"+
			"Email: %s\n\n"+
			"Сообщение:\n%s\n\n"+
			"—\nПерейдите на платформу для администраторов, чтобы ответить на сообщение клиента.\nКоманда Kroncl",
		data.TicketID, data.CompanyName, data.AccountName, data.AccountEmail, data.Message)

	// Создаём получателей
	recipients := make([]Recipient, len(data.AdminEmails))
	for i, email := range data.AdminEmails {
		recipients[i] = Recipient{Email: email}
	}

	msg := Message{
		Recipients: recipients,
		Subject:    subject,
		FromEmail:  s.config.NotifyDomain,
		FromName:   "Kroncl Support",
		Body: Body{
			HTML:      htmlBody,
			Plaintext: plainTextBody,
		},
		TrackLinks: 1,
		TrackRead:  1,
	}

	resp, err := s.Send(sendCtx, msg)
	if err != nil {
		log.Printf("❌ Failed to send support message to admins for ticket #%s: %v", data.TicketID, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Support message partially failed for ticket #%s: %v", data.TicketID, resp.FailedEmails)
	} else {
		log.Printf("✅ Support message sent to %d admins for ticket #%s from %s (%s)", len(data.AdminEmails), data.TicketID, data.CompanyName, data.AccountName)
	}
}
