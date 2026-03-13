package companies

import (
	"kroncl-server/internal/tenant/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool    *pgxpool.Pool
	storage *storage.Service
}

func NewService(pool *pgxpool.Pool, storage *storage.Service) *Service {
	return &Service{
		pool:    pool,
		storage: storage,
	}
}
