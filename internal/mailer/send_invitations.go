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
	CompanyName string
}

func (s *Service) SendCompanyInvitation(ctx context.Context, data *CompanyInvitationData) {
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	subject := fmt.Sprintf("Приглашение в компанию %s — Kroncl", data.CompanyName)

	baseDomain := config.GetClientDomain()
	inviteLink := baseDomain // пока похуй, просто ведём на основу

	htmlBody := fmt.Sprintf(`
		<h2>Приглашение в компанию</h2>
		<p>Здравствуйте!</p>
		<p>Вас приглашают присоединиться к компании <strong>%s</strong> на платформе Kroncl.</p>
		<p>Приняв приглашение, вы получите доступ к модулям компании в соответствии с назначенной вам ролью.</p>
		<p style="color: #666; font-size: 14px; margin: 20px 0;">Если у вас ещё нет аккаунта — после регистрации с этой почтой приглашение будет ждать вас в личном кабинете.</p>
		<div style="margin: 40px 0; text-align: center;">
			<a href="%s" style="display: inline-block; padding: 12px 28px; background-color: #e8551f; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 500; font-size: 15px;">Перейти в Kroncl</a>
		</div>
		<p style="color: #666; font-size: 14px;">Или перейдите по ссылке: <a href="%s" style="color: #e8551f;">%s</a></p>
		<p>Если вы не ожидали этого приглашения, просто проигнорируйте письмо.</p>
		<hr style="margin: 30px 0; border: none; border-top: 1px solid #e0e0e0;">
		<p style="color: #666; font-size: 12px;">С уважением,<br>Команда Kroncl</p>
	`, data.CompanyName, inviteLink, inviteLink, inviteLink)

	plainTextBody := fmt.Sprintf(
		"Приглашение в компанию — Kroncl\n\n"+
			"Здравствуйте!\n\n"+
			"Вас приглашают присоединиться к компании %s на платформе Kroncl.\n\n"+
			"Если у вас ещё нет аккаунта — после регистрации с этой почтой приглашение будет ждать вас в личном кабинете.\n\n"+
			"Перейдите по ссылке, чтобы принять приглашение:\n%s\n\n"+
			"Если вы не ожидали этого приглашения, просто проигнорируйте письмо.\n\n"+
			"—\nКоманда Kroncl",
		data.CompanyName, inviteLink)

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
