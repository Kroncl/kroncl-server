package mailer

import (
	"context"
	"fmt"
	"log"
	"time"
)

type ConfirmationCodeData struct {
	UserEmail string
	UserName  string
	Code      string
	ExpiresAt time.Time
}

func (s *Service) SendConfirmationCode(ctx context.Context, data *ConfirmationCodeData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := "Подтверждение регистрации — Kroncl"

	_, offset := data.ExpiresAt.Zone()
	offsetHours := offset / 3600
	timezone := fmt.Sprintf("UTC%+d", offsetHours)

	expiresTime := data.ExpiresAt.Format("02.01.2006 в 15:04:05") + " (" + timezone + ")"

	htmlBody := fmt.Sprintf(`
		<h2>Подтверждение регистрации</h2>
		<p>Здравствуйте, <strong>%s</strong>!</p>
		<p>Благодарим за регистрацию в Kroncl. Для завершения регистрации введите код подтверждения:</p>
		<div style="margin: 30px 0; padding: 20px; background-color: #f5f5f5; border-radius: 8px; text-align: center; border-bottom: 4px solid #e8551f;">
			<span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #333;">%s</span>
		</div>
		<p>Код действителен до: <strong>%s</strong>.</p>
		<p>Если вы не регистрировались в Kroncl, просто проигнорируйте это письмо.</p>
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">С уважением,<br>Команда Kroncl</p>
	`, data.UserName, data.Code, expiresTime)

	// Plain text версия для клиентов без HTML
	plainTextBody := fmt.Sprintf(
		"Подтверждение регистрации — Kroncl\n\n"+
			"Здравствуйте, %s!\n\n"+
			"Благодарим за регистрацию в Kroncl. Для завершения регистрации введите код подтверждения:\n\n"+
			"%s\n\n"+
			"Код действителен до: %s.\n\n"+
			"Если вы не регистрировались в Kroncl, просто проигнорируйте это письмо.\n\n"+
			"—\nКоманда Kroncl",
		data.UserName, data.Code, expiresTime)

	resp, err := s.SendSimple(sendCtx, data.UserEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to send confirmation code to %s: %v", data.UserEmail, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Confirmation code failed for %s: %v", data.UserEmail, resp.FailedEmails)
	} else {
		log.Printf("✅ Confirmation code sent to %s (expires at %s)", data.UserEmail, expiresTime)
	}
}

func (s *Service) SendConfirmationCodeResend(ctx context.Context, data *ConfirmationCodeData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := "Новый код подтверждения — Kroncl"

	_, offset := data.ExpiresAt.Zone()
	offsetHours := offset / 3600
	timezone := fmt.Sprintf("UTC%+d", offsetHours)
	expiresTime := data.ExpiresAt.Format("02.01.2006 в 15:04:05") + " (" + timezone + ")"

	htmlBody := fmt.Sprintf(`
		<h2>Новый код подтверждения</h2>
		<p>Здравствуйте, <strong>%s</strong>!</p>
		<p>Вы запросили новый код подтверждения для завершения регистрации в Kroncl.</p>
		<div style="margin: 30px 0; padding: 20px; background-color: #f5f5f5; border-radius: 8px; text-align: center; border-bottom: 4px solid #e8551f;">
			<span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #333;">%s</span>
		</div>
		<p>Код действителен до: <strong>%s</strong>.</p>
		<p>Если вы не запрашивали новый код, просто проигнорируйте это письмо.</p>
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">Команда Kroncl</p>
	`, data.UserName, data.Code, expiresTime)

	plainTextBody := fmt.Sprintf(
		"Новый код подтверждения — Kroncl\n\n"+
			"Здравствуйте, %s!\n\n"+
			"Вы запросили новый код подтверждения для завершения регистрации в Kroncl.\n\n"+
			"%s\n\n"+
			"Код действителен до: %s.\n\n"+
			"Если вы не запрашивали новый код, просто проигнорируйте это письмо.\n\n"+
			"—\nКоманда Kroncl",
		data.UserName, data.Code, expiresTime)

	resp, err := s.SendSimple(sendCtx, data.UserEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to resend confirmation code to %s: %v", data.UserEmail, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Confirmation code resend failed for %s: %v", data.UserEmail, resp.FailedEmails)
	} else {
		log.Printf("✅ Confirmation code resent to %s", data.UserEmail)
	}
}
