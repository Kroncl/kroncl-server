package companies

import (
	"kroncl-server/internal/mailer"
	"kroncl-server/internal/pricing"
	"kroncl-server/internal/tenant/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool           *pgxpool.Pool
	storage        *storage.Service
	pricingService *pricing.Service
	mailer         *mailer.Service
}

func NewService(
	pool *pgxpool.Pool,
	storage *storage.Service,
	pricingService *pricing.Service,
	mailer *mailer.Service,
) *Service {
	return &Service{
		pool:           pool,
		storage:        storage,
		pricingService: pricingService,
		mailer:         mailer,
	}
}

// переиспользование в permissioner
func (s *Service) GetPool() *pgxpool.Pool {
	return s.pool
}
