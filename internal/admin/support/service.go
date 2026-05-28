package adminsupport

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/mailer"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool             *pgxpool.Pool
	companiesService *companies.Service
	accountsService  *accounts.Service
	mailer           *mailer.Service
}

func NewService(
	pool *pgxpool.Pool,
	companiesService *companies.Service,
	accountsService *accounts.Service,
	mailer *mailer.Service,
) *Service {
	return &Service{
		pool:             pool,
		companiesService: companiesService,
		accountsService:  accountsService,
		mailer:           mailer,
	}
}
