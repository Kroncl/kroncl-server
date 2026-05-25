package support

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/mailer"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool             *pgxpool.Pool
	accountsService  *accounts.Service
	companiesService *companies.Service
	mailer           *mailer.Service
}

func NewService(pool *pgxpool.Pool, accountsService *accounts.Service, companiesService *companies.Service, mailer *mailer.Service) *Service {
	return &Service{pool: pool, accountsService: accountsService, companiesService: companiesService, mailer: mailer}
}
