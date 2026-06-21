package tgbot

import (
	"context"
	"kroncl-server/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mymmrac/telego"
)

type Service struct {
	pool   *pgxpool.Pool
	cfg    config.TelegramBotConfig
	domain string
	bot    *telego.Bot
}

func NewService(pool *pgxpool.Pool, cfg config.TelegramBotConfig) *Service {
	bot, err := telego.NewBot(cfg.Token)
	if err != nil {
		panic("failed to create telegram bot: " + err.Error())
	}

	s := &Service{
		pool:   pool,
		cfg:    cfg,
		domain: config.GetBaseDomain(),
		bot:    bot,
	}

	// Регистрируем вебхук
	webhookURL := "https://api." + s.domain + "/webhook/telegram/bot"
	if err := bot.SetWebhook(context.Background(), &telego.SetWebhookParams{
		URL:         webhookURL,
		SecretToken: cfg.WebhookSecret,
	}); err != nil {
		panic("failed to set webhook: " + err.Error())
	}

	return s
}
