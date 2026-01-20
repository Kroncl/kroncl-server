package companies

import (
	"context"
	"kroncl-server/internal/auth"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool       *pgxpool.Pool
	jwtService *auth.JWTService
}

func NewService(pool *pgxpool.Pool, jwtService *auth.JWTService) *Service {
	return &Service{
		pool:       pool,
		jwtService: jwtService,
	}
}

// func (s *Service) Create(slug string, name string, avatar_url string, is_public bool) (*Company, error) {

// }

func (s *Service) checkSlugUnique(slug string) (bool, error) {
	ctx := context.Background()
	var count int
	query := `SELECT COUNT(*) FROM companies WHERE slug = $1`

	err := s.pool.QueryRow(ctx, query, strings.ToLower(slug)).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
