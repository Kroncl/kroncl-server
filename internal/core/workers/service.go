package coreworkers

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/pricing"
	storagemedia "kroncl-server/internal/tenant/storage/media"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool                *pgxpool.Pool
	pricingService      *pricing.Service
	companiesService    *companies.Service
	accountsService     *accounts.Service
	storageMediaService *storagemedia.Service
}

func NewService(
	pool *pgxpool.Pool,
	pricingService *pricing.Service,
	companiesService *companies.Service,
	accountsService *accounts.Service,
	storageMediaService *storagemedia.Service,
) *Service {
	return &Service{
		pool:                pool,
		pricingService:      pricingService,
		companiesService:    companiesService,
		accountsService:     accountsService,
		storageMediaService: storageMediaService,
	}
}
