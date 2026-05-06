package adminsupport

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool             *pgxpool.Pool
	companiesService *companies.Service
	accountsService  *accounts.Service
}

func NewService(
	pool *pgxpool.Pool,
	companiesService *companies.Service,
	accountsService *accounts.Service,
) *Service {
	return &Service{
		pool:             pool,
		companiesService: companiesService,
		accountsService:  accountsService,
	}
}
