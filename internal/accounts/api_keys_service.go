package accounts

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) CreateApiKey(ctx context.Context, accountID string, req CreateApiKeyRequest) (*ApiKeyWithRaw, error) {
	// Проверка лимита ключей на аккаунт
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM api_keys WHERE account_id = $1 AND revoked_at IS NULL`, accountID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check api keys count: %w", err)
	}
	if count >= config.API_KEYS_PER_ACCOUNT {
		return nil, fmt.Errorf("maximum api keys limit reached (%d)", config.API_KEYS_PER_ACCOUNT)
	}

	rawKey, err := generateApiKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	keyHash, err := s.hashApiKey(rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	var expiresAt *time.Time
	if req.ExpiresIn != "" && req.ExpiresIn != "never" {
		duration, err := time.ParseDuration(req.ExpiresIn)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_in format (use 24h, 30d, etc): %w", err)
		}
		exp := time.Now().Add(duration)
		expiresAt = &exp
	}

	dailyRequests := config.API_KEY_DAILY_REQUESTS
	if req.DailyRequests != nil && *req.DailyRequests > 0 {
		dailyRequests = *req.DailyRequests
	}

	keyPrefix := config.API_KEY_PREFIX + rawKey[4:12]

	var key ApiKey
	query := `
		INSERT INTO api_keys (account_id, name, key_hash, key_prefix, daily_requests, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, account_id, name, key_prefix, daily_requests, expires_at, revoked_at, created_at, updated_at
	`

	err = s.pool.QueryRow(ctx, query,
		accountID, req.Name, keyHash, keyPrefix, dailyRequests, expiresAt,
	).Scan(
		&key.ID, &key.AccountID, &key.Name, &key.KeyPrefix,
		&key.DailyRequests, &key.ExpiresAt, &key.RevokedAt,
		&key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return &ApiKeyWithRaw{
		ApiKey: key,
		RawKey: rawKey,
	}, nil
}

func (s *Service) GetApiKey(ctx context.Context, accountID, keyID string) (*ApiKey, error) {
	var key ApiKey
	query := `
		SELECT id, account_id, name, key_prefix, daily_requests,
			last_used_at, expires_at, revoked_at, created_at, updated_at
		FROM api_keys
		WHERE id = $1 AND account_id = $2
	`

	err := s.pool.QueryRow(ctx, query, keyID, accountID).Scan(
		&key.ID, &key.AccountID, &key.Name, &key.KeyPrefix,
		&key.DailyRequests, &key.LastUsedAt, &key.ExpiresAt,
		&key.RevokedAt, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("api key not found: %w", err)
	}

	return &key, nil
}

func (s *Service) GetApiKeys(ctx context.Context, accountID string, req ApiKeyListRequest) (*ApiKeysResponse, error) {
	queryBase := `FROM api_keys`

	whereConditions := []string{`account_id = $1`}
	args := []interface{}{accountID}
	argIndex := 2

	if req.Status != nil && *req.Status != "" {
		switch *req.Status {
		case "active":
			whereConditions = append(whereConditions, "revoked_at IS NULL AND (expires_at IS NULL OR expires_at > NOW())")
		case "revoked":
			whereConditions = append(whereConditions, "revoked_at IS NOT NULL OR (expires_at IS NOT NULL AND expires_at <= NOW())")
		}
	}

	if req.Search != nil && *req.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf(
			"(name ILIKE $%d OR key_prefix ILIKE $%d)",
			argIndex, argIndex+1,
		))
		searchPattern := "%" + *req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	whereClause := " WHERE " + strings.Join(whereConditions, " AND ")

	countQuery := "SELECT COUNT(*) " + queryBase + whereClause
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count api keys: %w", err)
	}

	limit := 20
	if req.Limit > 0 {
		limit = req.Limit
	}
	if limit > 100 {
		limit = 100
	}

	page := 1
	if req.Page > 0 {
		page = req.Page
	}
	offset := (page - 1) * limit

	query := `
		SELECT 
			id, name, key_prefix, daily_requests,
			last_used_at, expires_at, revoked_at, created_at
	` + queryBase + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query api keys: %w", err)
	}
	defer rows.Close()

	var keys []ApiKeyListItem
	for rows.Next() {
		var k ApiKeyListItem
		err := rows.Scan(
			&k.ID, &k.Name, &k.KeyPrefix, &k.DailyRequests,
			&k.LastUsedAt, &k.ExpiresAt, &k.RevokedAt, &k.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}
		keys = append(keys, k)
	}

	pages := int(total) / limit
	if int(total)%limit > 0 {
		pages++
	}

	return &ApiKeysResponse{
		ApiKeys: keys,
		Pagination: core.Pagination{
			Total: int(total),
			Page:  page,
			Limit: limit,
			Pages: pages,
		},
	}, nil
}

func (s *Service) RevokeApiKey(ctx context.Context, accountID, keyID string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE api_keys 
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND account_id = $2 AND revoked_at IS NULL
	`, keyID, accountID)
	if err != nil {
		return fmt.Errorf("failed to revoke api key: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("api key not found or already revoked")
	}

	return nil
}

func (s *Service) ValidateApiKey(ctx context.Context, rawKey string) (*ApiKey, error) {
	query := `
		SELECT id, account_id, key_hash, revoked_at, expires_at, daily_requests
		FROM api_keys
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query api keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key ApiKey
		var keyHash string
		err := rows.Scan(&key.ID, &key.AccountID, &keyHash, &key.RevokedAt, &key.ExpiresAt, &key.DailyRequests)
		if err != nil {
			continue
		}

		if key.RevokedAt != nil {
			continue
		}
		if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
			continue
		}

		if s.verifyApiKey(keyHash, rawKey) {
			return &key, nil
		}
	}

	return nil, fmt.Errorf("invalid api key")
}

func (s *Service) UpdateApiKeyLastUsed(ctx context.Context, keyID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE api_keys SET last_used_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, keyID)
	return err
}

// --------
// UTILS
// --------

func generateApiKey() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return config.API_KEY_PREFIX + base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (s *Service) hashApiKey(key string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *Service) verifyApiKey(hash, key string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(key))
	return err == nil
}
