package tenantgeneral

import (
	"kroncl-server/internal/currency"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool            *pgxpool.Pool
	currencyService *currency.Service
}

func NewService(tenantPool *pgxpool.Pool, currencyService *currency.Service) *Service {
	return &Service{
		pool:            tenantPool,
		currencyService: currencyService,
	}
}
