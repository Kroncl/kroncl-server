package accounts

import (
	"kroncl-server/internal/auth"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/mailer"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool             *pgxpool.Pool
	jwtService       *auth.JWTService
	companiesService *companies.Service
	mailer           *mailer.Service
}

func NewService(
	pool *pgxpool.Pool,
	jwtService *auth.JWTService,
	companiesService *companies.Service,
	mailer *mailer.Service,
) *Service {
	return &Service{
		pool:             pool,
		jwtService:       jwtService,
		companiesService: companiesService,
		mailer:           mailer,
	}
}
