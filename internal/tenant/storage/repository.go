package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateStorageRecord(ctx context.Context, companyID string) (*Storage, error) {
	query := `
		INSERT INTO company_storage (company_id, schema_name, status)
		VALUES ($1, generate_tenant_schema_name($1), 'provisioning')
		RETURNING id, company_id, schema_name, status, storage_type, metadata, created_at, updated_at
	`

	var storage Storage
	err := r.pool.QueryRow(ctx, query, companyID).Scan(
		&storage.ID,
		&storage.CompanyID,
		&storage.SchemaName,
		&storage.Status,
		&storage.StorageType,
		&storage.Metadata,
		&storage.CreatedAt,
		&storage.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create storage record: %w", err)
	}

	return &storage, nil
}

func (r *Repository) GetStorageStatus(ctx context.Context, storageID string) (string, error) {
	query := `SELECT status FROM company_storage WHERE id = $1`

	var status string
	err := r.pool.QueryRow(ctx, query, storageID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("storage not found: %s", storageID)
		}
		return "", fmt.Errorf("failed to get storage status: %w", err)
	}

	return status, nil
}

func (r *Repository) UpdateStorageStatus(ctx context.Context, storageID, status string) error {
	query := `
		UPDATE company_storage 
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, status, storageID)
	if err != nil {
		return fmt.Errorf("failed to update storage status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	return nil
}

func (r *Repository) GetStorageByCompanyID(ctx context.Context, companyID string) (*Storage, error) {
	query := `
		SELECT id, company_id, schema_name, status, storage_type, metadata, created_at, updated_at
		FROM company_storage
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var storage Storage
	err := r.pool.QueryRow(ctx, query, companyID).Scan(
		&storage.ID,
		&storage.CompanyID,
		&storage.SchemaName,
		&storage.Status,
		&storage.StorageType,
		&storage.Metadata,
		&storage.CreatedAt,
		&storage.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get storage for company %s: %w", companyID, err)
	}

	return &storage, nil
}
