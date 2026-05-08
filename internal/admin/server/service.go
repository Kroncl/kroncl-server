package adminserver

import (
	coreworkers "kroncl-server/internal/core/workers"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool           *pgxpool.Pool
	metricsService *coreworkers.Service
}

func NewService(
	pool *pgxpool.Pool,
	metricsService *coreworkers.Service,
) *Service {
	return &Service{
		pool:           pool,
		metricsService: metricsService,
	}
}
