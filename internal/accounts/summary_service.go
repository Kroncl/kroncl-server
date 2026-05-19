package accounts

import (
	"context"
	"kroncl-server/internal/companies"
)

func (s *Service) GetAccountCompaniesCount(
	ctx context.Context,
	accountID string,
) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM company_accounts
		WHERE account_id = $1
	`, accountID).Scan(&count)

	return count, err
}

func (s *Service) GetPendingInvitationsCount(
	ctx context.Context,
	accountID string,
) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM company_invitations ci
		JOIN accounts a ON LOWER(a.email) = LOWER(ci.email)
		WHERE a.id = $1 AND ci.status = $2
	`, accountID, companies.InvitationStatusWaiting).Scan(&count)

	return count, err
}

func (s *Service) GetActiveFingerprintsCount(
	ctx context.Context,
	accountID string,
) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM account_fingerprints af
		JOIN fingerprints f ON af.fingerprint_id = f.id
		WHERE af.account_id = $1 AND f.status = 'active'
	`, accountID).Scan(&count)

	return count, err
}
