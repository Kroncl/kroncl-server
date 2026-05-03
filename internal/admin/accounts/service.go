package adminaccounts

import (
	"kroncl-server/internal/accounts"
	adminauth "kroncl-server/internal/admin/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool             *pgxpool.Pool
	accountsService  *accounts.Service
	adminAuthService *adminauth.Service
}

func NewService(
	pool *pgxpool.Pool,
	accountsService *accounts.Service,
	adminAuthService *adminauth.Service,
) *Service {
	return &Service{
		pool:             pool,
		accountsService:  accountsService,
		adminAuthService: adminAuthService,
	}
}
