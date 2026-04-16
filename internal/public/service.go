package public

import (
	"kroncl-server/internal/mailer"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool   *pgxpool.Pool
	mailer *mailer.Service
}

func NewService(
	pool *pgxpool.Pool,
	mailer *mailer.Service,
) *Service {
	return &Service{
		pool:   pool,
		mailer: mailer,
	}
}
