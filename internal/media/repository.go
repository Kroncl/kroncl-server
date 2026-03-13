package media

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

func (r *Repository) CreateFile(ctx context.Context, params CreateFileParams) (*File, error) {
	query := `
		INSERT INTO media_files (
			id, path, size, mime_type, created_by, original_name
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, path, size, mime_type, created_at, created_by, original_name
	`

	var file File
	err := r.pool.QueryRow(ctx, query,
		params.ID,
		params.Path,
		params.Size,
		params.MimeType,
		params.CreatedBy,
		params.OriginalName,
	).Scan(
		&file.ID,
		&file.Path,
		&file.Size,
		&file.MimeType,
		&file.CreatedAt,
		&file.CreatedBy,
		&file.OriginalName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return &file, nil
}

func (r *Repository) GetFile(ctx context.Context, id string) (*File, error) {
	query := `
		SELECT id, path, size, mime_type, created_at, created_by, original_name
		FROM media_files
		WHERE id = $1
	`

	var file File
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&file.ID,
		&file.Path,
		&file.Size,
		&file.MimeType,
		&file.CreatedAt,
		&file.CreatedBy,
		&file.OriginalName,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return &file, nil
}

func (r *Repository) DeleteFile(ctx context.Context, id string) error {
	query := `DELETE FROM media_files WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}
