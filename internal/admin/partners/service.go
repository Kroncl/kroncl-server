package adminpartners

import (
	"kroncl-server/internal/public"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool          *pgxpool.Pool
	publicService *public.Service
}

func NewService(
	pool *pgxpool.Pool,
	publicService *public.Service,
) *Service {
	return &Service{
		pool:          pool,
		publicService: publicService,
	}
}
