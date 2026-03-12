package dm

import (
	"kroncl-server/internal/tenant/fm"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool         *pgxpool.Pool
	fmRepository *fm.Repository
}

func NewRepository(pool *pgxpool.Pool, fmRepository *fm.Repository) *Repository {
	return &Repository{pool: pool, fmRepository: fmRepository}
}
