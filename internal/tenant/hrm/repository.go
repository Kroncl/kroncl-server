package hrm

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool             *pgxpool.Pool
	accountsService  *accounts.Service
	companiesService *companies.Service
}

func NewRepository(pool *pgxpool.Pool, accountsService *accounts.Service, companiesService *companies.Service) *Repository {
	return &Repository{pool: pool, accountsService: accountsService, companiesService: companiesService}
}
