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

// CreateFile сохраняет информацию о файле
func (r *Repository) CreateFile(ctx context.Context, params CreateFileParams) (*File, error) {
	query := `
		INSERT INTO media_files (
			path, url, size, mime_type, created_by, 
			original_name, metadata, entity_type, entity_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, path, url, size, mime_type, created_at, 
		          created_by, original_name, metadata, entity_type, entity_id
	`

	var file File
	err := r.pool.QueryRow(ctx, query,
		params.Path,
		params.URL,
		params.Size,
		params.MimeType,
		params.CreatedBy,
		params.OriginalName,
		params.Metadata,
		params.EntityType,
		params.EntityID,
	).Scan(
		&file.ID,
		&file.Path,
		&file.URL,
		&file.Size,
		&file.MimeType,
		&file.CreatedAt,
		&file.CreatedBy,
		&file.OriginalName,
		&file.Metadata,
		&file.EntityType,
		&file.EntityID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return &file, nil
}

// GetFile возвращает файл по ID
func (r *Repository) GetFile(ctx context.Context, id string) (*File, error) {
	query := `
		SELECT id, path, url, size, mime_type, created_at, 
		       created_by, original_name, metadata, entity_type, entity_id
		FROM media_files
		WHERE id = $1
	`

	var file File
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&file.ID,
		&file.Path,
		&file.URL,
		&file.Size,
		&file.MimeType,
		&file.CreatedAt,
		&file.CreatedBy,
		&file.OriginalName,
		&file.Metadata,
		&file.EntityType,
		&file.EntityID,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return &file, nil
}

// UpdateFileEntity обновляет привязку файла к сущности
func (r *Repository) UpdateFileEntity(ctx context.Context, id, entityType, entityID string) error {
	query := `
		UPDATE media_files 
		SET entity_type = $1, entity_id = $2
		WHERE id = $3
	`

	cmd, err := r.pool.Exec(ctx, query, entityType, entityID, id)
	if err != nil {
		return fmt.Errorf("failed to update file entity: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// DeleteFile удаляет запись о файле
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
