package mailer

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"log"
	"time"
)

type PasswordResetData struct {
	UserEmail string
	UserName  string
	Token     string
}

func (s *Service) SendPasswordReset(ctx context.Context, data *PasswordResetData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := "Сброс пароля — Kroncl"

	clientDomain := config.GetClientDomain()
	resetLink := fmt.Sprintf("%s/sso/recovery/reset-password?token=%s", clientDomain, data.Token)

	htmlBody := fmt.Sprintf(`
		<h2>Сброс пароля</h2>
		<p>Здравствуйте, <strong>%s</strong>!</p>
		<p>Вы запросили сброс пароля для вашего аккаунта в Kroncl.</p>
		<p>Для установки нового пароля нажмите на кнопку ниже:</p>
		<div style="margin: 40px 0; text-align: center;">
			<a href="%s" style="display: inline-block; padding: 12px 28px; background-color: #e8551f; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 500; font-size: 15px;">Сбросить пароль</a>
		</div>
		<p style="color: #666; font-size: 14px;">Или перейдите по ссылке: <a href="%s" style="color: #e8551f;">%s</a></p>
		<p style="color: #666; font-size: 14px;">Ссылка действительна в течение ограниченного времени.</p>
		<p>Если вы не запрашивали сброс пароля, просто проигнорируйте это письмо.</p>
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">С уважением,<br>Команда Kroncl</p>
	`, data.UserName, resetLink, resetLink, resetLink)

	plainTextBody := fmt.Sprintf(
		"Сброс пароля — Kroncl\n\n"+
			"Здравствуйте, %s!\n\n"+
			"Вы запросили сброс пароля для вашего аккаунта в Kroncl.\n\n"+
			"Для установки нового пароля перейдите по ссылке:\n%s\n\n"+
			"Ссылка действительна в течение ограниченного времени.\n\n"+
			"Если вы не запрашивали сброс пароля, просто проигнорируйте это письмо.\n\n"+
			"—\nКоманда Kroncl",
		data.UserName, resetLink)

	resp, err := s.SendSimple(sendCtx, data.UserEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to send password reset to %s: %v", data.UserEmail, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Password reset failed for %s: %v", data.UserEmail, resp.FailedEmails)
	} else {
		log.Printf("✅ Password reset sent to %s", data.UserEmail)
	}
}
