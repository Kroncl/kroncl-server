package corestatus

import (
	coreworkers "kroncl-server/internal/core/workers"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool        *pgxpool.Pool
	coreWorkers *coreworkers.Service
}

func NewService(
	pool *pgxpool.Pool,
	coreWorkers *coreworkers.Service,
) *Service {
	return &Service{
		pool:        pool,
		coreWorkers: coreWorkers,
	}
}
