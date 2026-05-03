package admindb

import (
	coreworkers "kroncl-server/internal/core/workers"
	"kroncl-server/internal/migrator"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool           *pgxpool.Pool
	metricsService *coreworkers.Service
	migrator       *migrator.Migrator
}

func NewService(
	pool *pgxpool.Pool,
	metricsService *coreworkers.Service,
	migrator *migrator.Migrator,
) *Service {
	return &Service{
		pool:           pool,
		metricsService: metricsService,
		migrator:       migrator,
	}
}
