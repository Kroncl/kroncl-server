package public

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/mailer"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) countWaitingByEmail(ctx context.Context, email string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM incoming_partners WHERE email = $1 AND status = $2`
	err := s.pool.QueryRow(ctx, query, strings.ToLower(email), PARTNER_STATUS_WAITING).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count waiting partners: %w", err)
	}
	return count, nil
}

func (s *Service) Create(ctx context.Context, req CreateIncomingPartnerRequest) (bool, error) {
	if req.Name == "" {
		return false, fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return false, fmt.Errorf("email is required")
	}
	if req.Type != PARTNER_TYPE_PUBLIC && req.Type != PARTNER_TYPE_PRIVATE {
		return false, fmt.Errorf("invalid partner type")
	}

	count, err := s.countWaitingByEmail(ctx, req.Email)
	if err != nil {
		return false, err
	}
	if count >= config.PARTNERS_SAME_MAX_IN_ROW {
		// всё равно говорим, что всё заебись
		return true, nil
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return false, fmt.Errorf("failed to generate UUID: %w", err)
	}

	currentTime := time.Now()

	query := `
		INSERT INTO incoming_partners (id, name, type, text, email, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, type, text, email, status, created_at, updated_at
	`

	var partner IncomingPartner
	var createdAt, updatedAt time.Time

	err = s.pool.QueryRow(
		ctx,
		query,
		id.String(),
		req.Name,
		req.Type,
		req.Text,
		strings.ToLower(req.Email),
		PARTNER_STATUS_WAITING,
		currentTime,
		currentTime,
	).Scan(
		&partner.ID,
		&partner.Name,
		&partner.Type,
		&partner.Text,
		&partner.Email,
		&partner.Status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return false, fmt.Errorf("failed to create partner: %w", err)
	}

	go func() {
		data := &mailer.BecomePartnerData{
			CompanyEmail: req.Email,
			CompanyName:  req.Name,
		}
		s.mailer.SendBecomePartnerRequest(context.Background(), data)
	}()

	return true, nil
}
