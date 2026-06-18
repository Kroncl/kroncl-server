package billing

import (
	"kroncl-server/internal/pricing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rentifly/tinkoff"
)

type Service struct {
	pool           *pgxpool.Pool
	tbankClient    *tinkoff.Client
	webhookURL     string
	pricingService *pricing.Service
}

func NewService(
	pool *pgxpool.Pool,
	tbankClient *tinkoff.Client,
	webhookBaseURL string,
	pricingService *pricing.Service,
) *Service {
	return &Service{
		pool:           pool,
		tbankClient:    tbankClient,
		webhookURL:     webhookBaseURL,
		pricingService: pricingService,
	}
}

func (s *Service) GetWebhookURL() string {
	return s.webhookURL
}
