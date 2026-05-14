package adminpricing

import (
	"kroncl-server/internal/pricing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool           *pgxpool.Pool
	pricingService *pricing.Service
}

func NewService(
	pool *pgxpool.Pool,
	pricingService *pricing.Service,
) *Service {
	return &Service{
		pool:           pool,
		pricingService: pricingService,
	}
}
