package adminaccounts

import (
	"kroncl-server/internal/accounts"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool            *pgxpool.Pool
	accountsService *accounts.Service
}

func NewService(
	pool *pgxpool.Pool,
	accountsService *accounts.Service,
) *Service {
	return &Service{
		pool:            pool,
		accountsService: accountsService,
	}
}
