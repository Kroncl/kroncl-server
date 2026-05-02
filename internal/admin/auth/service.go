package adminauth

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(
	pool *pgxpool.Pool,
) *Service {
	return &Service{
		pool: pool,
	}
}
