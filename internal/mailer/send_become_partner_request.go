package mailer

import (
	"context"
	"fmt"
	"log"
	"time"
)

type BecomePartnerData struct {
	CompanyEmail string
	CompanyName  string
}

func (s *Service) SendBecomePartnerRequest(ctx context.Context, data *BecomePartnerData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := "Заявка на партнёрство — Kroncl"

	htmlBody := fmt.Sprintf(`
		<h2>Заявка получена</h2>
		<p>Здравствуйте, <strong>%s</strong>!</p>
		<p>Ваша заявка на партнёрство с Kroncl получена и находится в обработке.</p>
		<p>Мы свяжемся с вами в ближайшее время для обсуждения деталей сотрудничества.</p>
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">С уважением,<br>Команда Kroncl</p>
	`, data.CompanyName)

	plainTextBody := fmt.Sprintf(
		"Заявка на партнёрство — Kroncl\n\n"+
			"Здравствуйте, %s!\n\n"+
			"Ваша заявка на партнёрство с Kroncl получена и находится в обработке.\n\n"+
			"Мы свяжемся с вами в ближайшее время для обсуждения деталей сотрудничества.\n\n"+
			"—\nКоманда Kroncl",
		data.CompanyName)

	resp, err := s.SendSimple(sendCtx, data.CompanyEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to send become partner confirmation to %s: %v", data.CompanyEmail, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Become partner confirmation failed for %s: %v", data.CompanyEmail, resp.FailedEmails)
	} else {
		log.Printf("✅ Become partner confirmation sent to %s", data.CompanyEmail)
	}
}
