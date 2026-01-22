package hrm

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*CreateEmployeeResponse, error) {
	return nil, nil
}
