package mailer

import (
	"context"
	"fmt"
	"log"
	"time"
)

type PasswordResetSuccessData struct {
	UserEmail string
	UserName  string
}

func (s *Service) SendPasswordResetSuccess(ctx context.Context, data *PasswordResetSuccessData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := "Пароль успешно изменён — Kroncl"

	htmlBody := fmt.Sprintf(`
		<h2>Пароль изменён</h2>
		<p>Здравствуйте, <strong>%s</strong>!</p>
		<p>Пароль для вашего аккаунта в Kroncl был успешно изменён.</p>
		<p>Если вы не выполняли это действие, немедленно свяжитесь с поддержкой.</p>
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">С уважением,<br>Команда Kroncl</p>
	`, data.UserName)

	plainTextBody := fmt.Sprintf(
		"Пароль успешно изменён — Kroncl\n\n"+
			"Здравствуйте, %s!\n\n"+
			"Пароль для вашего аккаунта в Kroncl был успешно изменён.\n\n"+
			"Если вы не выполняли это действие, немедленно свяжитесь с поддержкой.\n\n"+
			"—\nКоманда Kroncl",
		data.UserName)

	resp, err := s.SendSimple(sendCtx, data.UserEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to send password reset success to %s: %v", data.UserEmail, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Password reset success failed for %s: %v", data.UserEmail, resp.FailedEmails)
	} else {
		log.Printf("✅ Password reset success sent to %s", data.UserEmail)
	}
}
