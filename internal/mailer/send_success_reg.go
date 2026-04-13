package mailer

import (
	"context"
	"fmt"
	"log"
	"time"
)

type RegistrationSuccessData struct {
	UserEmail string
	UserName  string
}

func (s *Service) SendRegistrationSuccess(ctx context.Context, data *RegistrationSuccessData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := "Регистрация завершена — Kroncl"

	htmlBody := fmt.Sprintf(`
		<h2>Регистрация завершена</h2>
		<p>Здравствуйте, <strong>%s</strong>!</p>
		<p>Ваш аккаунт в Kroncl успешно подтверждён и готов к работе.</p>
		<p>Теперь вы можете создавать компании, приглашать сотрудников и управлять бизнесом в единой системе.</p>
		<p>Для входа используйте свой email и пароль.</p>
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">С уважением,<br>Команда Kroncl</p>
	`, data.UserName)

	plainTextBody := fmt.Sprintf(
		"Регистрация завершена — Kroncl\n\n"+
			"Здравствуйте, %s!\n\n"+
			"Ваш аккаунт в Kroncl успешно подтверждён и готов к работе.\n\n"+
			"Теперь вы можете создавать компании, приглашать сотрудников и управлять бизнесом в единой системе.\n\n"+
			"Для входа используйте свой email и пароль.\n\n"+
			"—\nКоманда Kroncl",
		data.UserName)

	resp, err := s.SendSimple(sendCtx, data.UserEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to send registration success to %s: %v", data.UserEmail, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Registration success failed for %s: %v", data.UserEmail, resp.FailedEmails)
	} else {
		log.Printf("✅ Registration success sent to %s", data.UserEmail)
	}
}
