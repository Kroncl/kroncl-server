package mailer

import (
	"context"
	"fmt"
	"log"
	"time"
)

// LoginNotificationData данные для уведомления о входе в аккаунт
type LoginNotificationData struct {
	UserEmail string
	UserName  string
	IPAddress string
	LoginTime time.Time
}

// SendLoginNotification отправляет уведомление о входе в аккаунт (запускается в горутине)
func (s *Service) SendLoginNotification(ctx context.Context, data *LoginNotificationData) {
	// Используем отдельный контекст с таймаутом
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Формируем письмо
	subject := "Новый вход в аккаунт"

	// Форматируем время с указанием часового пояса
	// Получаем смещение в часах
	_, offset := data.LoginTime.Zone()
	offsetHours := offset / 3600
	timezone := fmt.Sprintf("UTC%+d", offsetHours)

	loginTime := data.LoginTime.Format("02.01.2006 в 15:04:05") + " (" + timezone + ")"

	htmlBody := fmt.Sprintf("<h2>Новый вход в аккаунт</h2><p>Здравствуйте, <strong>%s</strong>!</p><p>Вход выполнен:</p><ul><li>Время: %s</li><li>IP: %s</li></ul><p>Если это были не вы, смените пароль.</p>",
		data.UserName, loginTime, data.IPAddress)

	plainTextBody := fmt.Sprintf("Новый вход в аккаунт\n\nЗдравствуйте, %s!\n\nВход выполнен:\nВремя: %s\nIP: %s\n\nЕсли это были не вы, смените пароль.",
		data.UserName, loginTime, data.IPAddress)

	// Отправляем
	resp, err := s.SendSimple(sendCtx, data.UserEmail, subject, htmlBody, plainTextBody)
	if err != nil {
		log.Printf("❌ Failed to send login notification to %s: %v", data.UserEmail, err)
		return
	}

	if len(resp.FailedEmails) > 0 {
		log.Printf("⚠️ Login notification failed for %s: %v", data.UserEmail, resp.FailedEmails)
	} else {
		log.Printf("✅ Login notification sent to %s", data.UserEmail)
	}
}
