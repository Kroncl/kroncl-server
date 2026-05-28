package mailer

import (
	"context"
	"fmt"
	"log"
	"time"
)

type ClientSupportMessageData struct {
	ClientEmail string
	ClientName  string
	Message     string
	TicketID    string
}

func (s *Service) SendSupportAnswer(ctx context.Context, data *ClientSupportMessageData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := fmt.Sprintf("[Тикет #%s] | Ответ на обращение в поддержку Kroncl", data.TicketID)

	htmlBody := fmt.Sprintf(`
		<h2>Ответ на ваше обращение в техническую поддержку</h2>
		
		<p><strong>Тикет:</strong> #%s</p>
		
		<p><strong>Сообщение от поддержки:</strong><br>
		<div style="background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin: 10px 0;">
			%s
		</div>
		</p>
		
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">Вы можете продолжить общение в панели управления вашей компанией.<br>
		С уважением,<br>Команда Kroncl</p>
	`, data.TicketID, data.Message)

	plainTextBody := fmt.Sprintf(
		"Ответ на ваше обращение в техническую поддержку Kroncl\n\n"+
			"Тикет: #%s\n\n"+
			"Сообщение от поддержки:\n%s\n\n"+
			"—\nВы можете продолжить общение в панели управления вашей компанией.\nКоманда Kroncl",
		data.TicketID, data.Message)

	resp, err := s.SendSimple(sendCtx, data.ClientEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to send support reply to client %s for ticket #%s: %v", data.ClientEmail, data.TicketID, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Support reply failed for client %s ticket #%s: %v", data.ClientEmail, data.TicketID, resp.FailedEmails)
	} else {
		log.Printf("✅ Support reply sent to client %s for ticket #%s", data.ClientEmail, data.TicketID)
	}
}
