package permissioner

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
	return &Service{
		pool: pool,
	}
}

func (s *Service) Has(
	ctx context.Context,
	userID, companyID, permission string,
) (bool, error) {
	userPerms, err := s.getUserPermissions(ctx, userID, companyID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return s.checkPermission(userPerms, permission), nil
}

func (s *Service) HasAll(
	ctx context.Context,
	userID, companyID string,
	permissions ...string,
) (bool, error) {
	if len(permissions) == 0 {
		return true, nil
	}

	userPerms, err := s.getUserPermissions(ctx, userID, companyID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if !s.checkPermission(userPerms, perm) {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) HasAny(
	ctx context.Context,
	userID, companyID string,
	permissions ...string,
) (bool, error) {
	if len(permissions) == 0 {
		return true, nil
	}

	userPerms, err := s.getUserPermissions(ctx, userID, companyID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if s.checkPermission(userPerms, perm) {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) getUserPermissions(
	ctx context.Context,
	userID, companyID string,
) (map[string]bool, error) {
	query := `
		SELECT 
			COALESCE(r.permissions, '[]'::jsonb) as role_perms,
			COALESCE(ca.permissions, '{}'::jsonb) as custom_perms
		FROM company_accounts ca
		LEFT JOIN roles r ON ca.role_id = r.id
		WHERE ca.company_id = $1 AND ca.account_id = $2
	`

	var (
		rolePerms   []string
		customPerms map[string]bool
	)

	err := s.pool.QueryRow(ctx, query, companyID, userID).Scan(&rolePerms, &customPerms)
	if err != nil {
		return nil, fmt.Errorf("user not found in company: %w", err)
	}

	return mergePermissions(rolePerms, customPerms), nil
}

func mergePermissions(rolePerms []string, customPerms map[string]bool) map[string]bool {
	result := make(map[string]bool)

	// Права из роли
	for _, perm := range rolePerms {
		result[perm] = true
	}

	// Кастомные оверрайды (только добавляют)
	for perm, allowed := range customPerms {
		if allowed {
			result[perm] = true
		}
		// false игнорируем - не удаляем права роли
	}

	return result
}

func (s *Service) checkPermission(userPerms map[string]bool, permission string) bool {
	// Полный доступ
	if userPerms["*"] {
		return true
	}

	// Точное совпадение
	if allowed, ok := userPerms[permission]; ok {
		return allowed
	}

	// Wildcard проверка (crm.* для crm.clients.view)
	parts := strings.Split(permission, ".")
	for i := 1; i < len(parts); i++ {
		wildcard := strings.Join(parts[:i], ".") + ".*"
		if allowed, ok := userPerms[wildcard]; ok {
			return allowed
		}
	}

	return false
}
