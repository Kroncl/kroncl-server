package adminauth

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"

	"github.com/jackc/pgx/v5"
)

func (s *Service) PromoteToAdmin(ctx context.Context, accountID string, level int) error {
	if level < config.ADMIN_LEVEL_MIN || level > config.ADMIN_LEVEL_MAX {
		return fmt.Errorf("level must be between %d and %d", config.ADMIN_LEVEL_MIN, config.ADMIN_LEVEL_MAX)
	}

	query := `
		INSERT INTO admins (account_id, level)
		VALUES ($1, $2)
		ON CONFLICT (account_id) DO UPDATE SET
			level = EXCLUDED.level,
			updated_at = NOW()
	`

	_, err := s.pool.Exec(ctx, query, accountID, level)
	if err != nil {
		return fmt.Errorf("failed to promote account to admin: %w", err)
	}

	return nil
}

func (s *Service) DemoteFromAdmin(ctx context.Context, accountID string) error {
	query := `DELETE FROM admins WHERE account_id = $1`

	_, err := s.pool.Exec(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("failed to demote account from admin: %w", err)
	}

	return nil
}

func (s *Service) IsAdmin(ctx context.Context, accountID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM admins WHERE account_id = $1)`

	var exists bool
	err := s.pool.QueryRow(ctx, query, accountID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check admin status: %w", err)
	}

	return exists, nil
}

func (s *Service) GetAdminLevel(ctx context.Context, accountID string) (int, error) {
	query := `SELECT level FROM admins WHERE account_id = $1`

	var level int
	err := s.pool.QueryRow(ctx, query, accountID).Scan(&level)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get admin level: %w", err)
	}

	return level, nil
}
