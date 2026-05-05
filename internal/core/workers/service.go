package coreworkers

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/pricing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool             *pgxpool.Pool
	pricingService   *pricing.Service
	companiesService *companies.Service
	accountsService  *accounts.Service
}

func NewService(
	pool *pgxpool.Pool,
	pricingService *pricing.Service,
	companiesService *companies.Service,
	accountsService *accounts.Service,
) *Service {
	return &Service{
		pool:             pool,
		pricingService:   pricingService,
		companiesService: companiesService,
		accountsService:  accountsService,
	}
}
