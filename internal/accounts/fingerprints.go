package accounts

import (
	"context"
	"encoding/base64"
	"fmt"
	"kroncl-server/internal/core"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateFingerprint создает новый фингерпринт для аккаунта
func (s *Service) CreateFingerprint(ctx context.Context, accountID string, expiresIn *string) (*FingerprintWithKey, error) {
	// Генерируем ключ
	key, err := generateFingerprintKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Хэшируем ключ
	hash, err := s.hashFingerprint(key)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	// Считаем expiration
	var expiredAt *time.Time
	if expiresIn != nil && *expiresIn != "never" && *expiresIn != "" {
		duration, err := time.ParseDuration(*expiresIn)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_in format (use 24h, 30d, etc): %w", err)
		}
		exp := time.Now().Add(duration)
		expiredAt = &exp
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Создаем фингерпринт
	var fp Fingerprint
	var fpID uuid.UUID

	query := `
        INSERT INTO fingerprints (hash, status, expired_at, created_at)
        VALUES ($1, 'active', $2, NOW())
        RETURNING id, status, expired_at, created_at
    `

	err = tx.QueryRow(ctx, query, hash, expiredAt).Scan(
		&fpID, &fp.Status, &fp.ExpiredAt, &fp.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fingerprint: %w", err)
	}
	fp.ID = fpID.String()

	// Связываем с аккаунтом
	_, err = tx.Exec(ctx, `
        INSERT INTO account_fingerprints (account_id, fingerprint_id, created_at)
        VALUES ($1, $2, NOW())
    `, accountID, fpID)
	if err != nil {
		return nil, fmt.Errorf("failed to link fingerprint to account: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Возвращаем ключ только один раз!
	return &FingerprintWithKey{
		Fingerprint: fp,
		Key:         key,
	}, nil
}

// LoginWithFingerprint вход по фингерпринту
func (s *Service) LoginWithFingerprint(ctx context.Context, key string) (accessToken, refreshToken string, account *Account, err error) {
	// Достаём все фингерпринты (или используем частичный индекс)
	query := `
        SELECT 
            f.id, f.hash, f.status, f.expired_at, 
            af.last_used_at, a.id as account_id
        FROM fingerprints f
        JOIN account_fingerprints af ON f.id = af.fingerprint_id
        JOIN accounts a ON af.account_id = a.id
        WHERE f.status = 'active' 
          AND (f.expired_at IS NULL OR f.expired_at > NOW())
    `

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return "", "", nil, fmt.Errorf("login failed: %w", err)
	}
	defer rows.Close()

	// Ищем подходящий ключ
	var accountID string
	var fpID uuid.UUID
	var found bool

	for rows.Next() {
		var hash string
		var status string
		var expiredAt *time.Time
		var lastUsedAt *time.Time
		var accID string
		var id uuid.UUID

		err := rows.Scan(&id, &hash, &status, &expiredAt, &lastUsedAt, &accID)
		if err != nil {
			continue
		}

		// Проверяем ключ
		if s.verifyFingerprint(hash, key) {
			accountID = accID
			fpID = id
			found = true
			break
		}
	}

	if !found {
		return "", "", nil, fmt.Errorf("invalid fingerprint key")
	}

	// Обновляем last_used_at
	_, err = s.pool.Exec(ctx, `
        UPDATE account_fingerprints 
        SET last_used_at = NOW() 
        WHERE fingerprint_id = $1
    `, fpID)
	if err != nil {
		log.Printf("failed to update last_used_at: %v", err)
	}

	// Получаем аккаунт для проверки статуса
	account, err = s.GetByID(ctx, accountID)
	if err != nil {
		return "", "", nil, fmt.Errorf("account not found")
	}

	if account.Status != "confirmed" {
		return "", "", nil, fmt.Errorf("account not confirmed")
	}

	// Генерируем токены
	accessToken, err = s.jwtService.GenerateAccessToken(account.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = s.jwtService.GenerateRefreshToken(account.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, account, nil
}

func (s *Service) verifyFingerprint(hash, key string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(key))
	return err == nil
}

// RevokeFingerprint отзывает фингерпринт
func (s *Service) RevokeFingerprint(ctx context.Context, accountID, fingerprintID string) error {
	// Проверяем, что фингерпринт принадлежит этому аккаунту
	var exists bool
	err := s.pool.QueryRow(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM account_fingerprints 
            WHERE account_id = $1 AND fingerprint_id = $2
        )
    `, accountID, fingerprintID).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to verify ownership: %w", err)
	}

	if !exists {
		return fmt.Errorf("fingerprint not found or does not belong to you")
	}

	// Отзываем (мягкое удаление - меняем статус)
	_, err = s.pool.Exec(ctx, `
        UPDATE fingerprints 
        SET status = 'inactive' 
        WHERE id = $1
    `, fingerprintID)

	if err != nil {
		return fmt.Errorf("failed to revoke fingerprint: %w", err)
	}

	return nil
}

// GetAccountFingerprints возвращает список фингерпринтов аккаунта с пагинацией и фильтрацией
func (s *Service) GetAccountFingerprints(ctx context.Context, accountID string, req FingerprintListRequest) (*FingerprintsResponse, error) {
	// Базовый запрос
	queryBase := `FROM fingerprints f
                  JOIN account_fingerprints af ON f.id = af.fingerprint_id`

	// WHERE условия
	whereConditions := []string{`af.account_id = $1`}
	args := []interface{}{accountID}
	argIndex := 2

	if req.Status != nil && *req.Status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("f.status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		// Поиск по ID или по маске (используем last 8 символов хэша для маски)
		whereConditions = append(whereConditions, fmt.Sprintf(
			"(f.id::text ILIKE $%d OR RIGHT(f.hash, 8) ILIKE $%d)",
			argIndex, argIndex+1,
		))
		searchPattern := "%" + *req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	whereClause := " WHERE " + strings.Join(whereConditions, " AND ")

	// Общее количество
	countQuery := "SELECT COUNT(*) " + queryBase + whereClause
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count fingerprints: %w", err)
	}

	// Пагинация
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

	// Основной запрос
	query := `
        SELECT 
            f.id, f.status, f.expired_at, f.created_at,
            af.last_used_at, RIGHT(f.hash, 8) as hash_suffix
    ` + queryBase + whereClause + `
        ORDER BY f.created_at DESC
        LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query fingerprints: %w", err)
	}
	defer rows.Close()

	var fingerprints []FingerprintListItem
	for rows.Next() {
		var fp FingerprintListItem
		var hashSuffix string
		var id uuid.UUID

		err := rows.Scan(
			&id,
			&fp.Status,
			&fp.ExpiredAt,
			&fp.CreatedAt,
			&fp.LastUsedAt,
			&hashSuffix,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fingerprint: %w", err)
		}

		fp.ID = id.String()
		fp.MaskedKey = "fp_... " + hashSuffix // или другой формат маски

		fingerprints = append(fingerprints, fp)
	}

	// Считаем страницы
	pages := int(total) / limit
	if int(total)%limit > 0 {
		pages++
	}

	pagination := &core.Pagination{
		Total: int(total),
		Page:  page,
		Limit: limit,
		Pages: pages,
	}

	return &FingerprintsResponse{
		Fingerprints: fingerprints,
		Pagination:   *pagination,
	}, nil
}

// Вспомогательные функции
func generateFingerprintKey() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	key := base64.RawURLEncoding.EncodeToString(bytes)
	return "fp_" + key, nil
}

func (s *Service) hashFingerprint(key string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
