package companies

import (
	"context"
	"fmt"
)

func (s *Service) Drop(ctx context.Context, companyID string) error {
	exists, err := s.checkCompanyExists(ctx, companyID)
	if err != nil {
		return fmt.Errorf("failed to check company existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("company not found")
	}

	if err := s.storage.DropStorage(ctx, companyID); err != nil {
		return fmt.Errorf("failed to drop storage: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM company_accounts WHERE company_id = $1`, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete company members: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM companies WHERE id = $1`, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete company: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) checkCompanyExists(ctx context.Context, companyID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`
	err := s.pool.QueryRow(ctx, query, companyID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
