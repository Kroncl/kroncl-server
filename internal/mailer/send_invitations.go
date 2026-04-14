package mailer

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"log"
	"time"
)

type CompanyInvitationData struct {
	UserEmail   string
	UserName    string
	CompanyName string
	InviterName string
}

func (s *Service) SendCompanyInvitation(ctx context.Context, data *CompanyInvitationData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := fmt.Sprintf("Приглашение в компанию %s — Kroncl", data.CompanyName)

	baseDomain := config.GetBaseDomain()
	inviteLink := baseDomain // пока похуй ведём просто на главную, дальше логин и просмотр приглашений

	htmlBody := fmt.Sprintf(`
		<h2>Приглашение в компанию</h2>
		<p>Здравствуйте, <strong>%s</strong>!</p>
		<p><strong>%s</strong> приглашает вас присоединиться к компании <strong>%s</strong> на платформе Kroncl.</p>
		<p>Приняв приглашение, вы получите доступ к модулям компании в соответствии с назначенной вам ролью.</p>
		<div style="margin: 40px 0; text-align: center;">
			<a href="%s" style="display: inline-block; padding: 14px 32px; background-color: #e8551f; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 500; font-size: 16px;">Принять приглашение</a>
		</div>
		<p style="color: #666; font-size: 14px;">Или перейдите по ссылке: <a href="%s" style="color: #e8551f;">%s</a></p>
		<p>Если вы не ожидали этого приглашения, просто проигнорируйте письмо.</p>
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">С уважением,<br>Команда Kroncl</p>
	`, data.UserName, data.InviterName, data.CompanyName, inviteLink, inviteLink, inviteLink)

	plainTextBody := fmt.Sprintf(
		"Приглашение в компанию — Kroncl\n\n"+
			"Здравствуйте, %s!\n\n"+
			"%s приглашает вас присоединиться к компании %s на платформе Kroncl.\n\n"+
			"Приняв приглашение, вы получите доступ к модулям компании в соответствии с назначенными вам разрешениям.\n\n"+
			"Перейдите по ссылке, чтобы принять приглашение:\n%s\n\n"+
			"Если вы не ожидали этого приглашения, просто проигнорируйте письмо.\n\n"+
			"—\nКоманда Kroncl",
		data.UserName, data.InviterName, data.CompanyName, inviteLink)

	resp, err := s.SendSimple(sendCtx, data.UserEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to send company invitation to %s: %v", data.UserEmail, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Company invitation failed for %s: %v", data.UserEmail, resp.FailedEmails)
	} else {
		log.Printf("✅ Company invitation sent to %s for company %s", data.UserEmail, data.CompanyName)
	}
}
