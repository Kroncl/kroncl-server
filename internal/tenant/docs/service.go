package docs

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) CreateDoc(ctx context.Context, req CreateDocRequest) (*Doc, error) {
	query := `
		INSERT INTO docs (object_path, module, type, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, object_path, module, type, comment, created_at, updated_at
	`

	var doc Doc
	err := s.pool.QueryRow(ctx, query, req.ObjectPath, req.Module, req.Type, req.Comment).Scan(
		&doc.ID,
		&doc.ObjectPath,
		&doc.Module,
		&doc.Type,
		&doc.Comment,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create doc: %w", err)
	}

	return &doc, nil
}

func (s *Service) GetDocs(ctx context.Context, offset, limit int, module, docType, search *string) ([]Doc, int64, error) {
	var args []interface{}
	var whereConditions []string
	argIndex := 1

	if module != nil && *module != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("module = $%d", argIndex))
		args = append(args, *module)
		argIndex++
	}

	if docType != nil && *docType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *docType)
		argIndex++
	}

	if search != nil && *search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("comment ILIKE $%d", argIndex))
		args = append(args, "%"+*search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM docs " + whereClause
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count docs: %w", err)
	}

	query := `
		SELECT id, object_path, module, type, comment, created_at, updated_at
		FROM docs
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get docs: %w", err)
	}
	defer rows.Close()

	var docs []Doc
	for rows.Next() {
		var doc Doc
		err := rows.Scan(
			&doc.ID,
			&doc.ObjectPath,
			&doc.Module,
			&doc.Type,
			&doc.Comment,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan doc: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, total, nil
}

func (s *Service) GetDocByID(ctx context.Context, id string) (*Doc, error) {
	query := `
		SELECT id, object_path, module, type, comment, created_at, updated_at
		FROM docs
		WHERE id = $1
	`

	var doc Doc
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID,
		&doc.ObjectPath,
		&doc.Module,
		&doc.Type,
		&doc.Comment,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get doc: %w", err)
	}

	return &doc, nil
}
