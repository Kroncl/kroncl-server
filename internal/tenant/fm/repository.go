package fm

import (
	"kroncl-server/internal/tenant/hrm"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool                *pgxpool.Pool
	employeesRepository *hrm.Repository
}

func NewRepository(pool *pgxpool.Pool, employeesRepository *hrm.Repository) *Repository {
	return &Repository{pool: pool, employeesRepository: employeesRepository}
}
