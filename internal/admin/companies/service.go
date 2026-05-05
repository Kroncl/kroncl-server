package admincompanies

import (
	"kroncl-server/internal/companies"
	"kroncl-server/internal/tenant/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool             *pgxpool.Pool
	companiesService *companies.Service
	storageService   *storage.Service
}

func NewService(
	pool *pgxpool.Pool,
	companiesService *companies.Service,
	storageService *storage.Service,
) *Service {
	return &Service{
		pool:             pool,
		companiesService: companiesService,
		storageService:   storageService,
	}
}
