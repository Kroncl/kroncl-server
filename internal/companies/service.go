package companies

import (
	"kroncl-server/internal/billing"
	"kroncl-server/internal/mailer"
	"kroncl-server/internal/pricing"
	"kroncl-server/internal/tenant/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool           *pgxpool.Pool
	storageService *storage.Service
	pricingService *pricing.Service
	mailer         *mailer.Service
	billingService *billing.Service
}

func NewService(
	pool *pgxpool.Pool,
	storageService *storage.Service,
	pricingService *pricing.Service,
	mailer *mailer.Service,
	billingService *billing.Service,
) *Service {
	return &Service{
		pool:           pool,
		storageService: storageService,
		pricingService: pricingService,
		mailer:         mailer,
		billingService: billingService,
	}
}

// переиспользование в permissioner
func (s *Service) GetPool() *pgxpool.Pool {
	return s.pool
}
