package permissioner

import (
	"context"
	"encoding/json"
	"fmt"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	companiesService *companies.Service
}

func NewService(companiesService *companies.Service) *Service {
	return &Service{
		companiesService: companiesService,
	}
}

// PermissionCheckResult содержит детальную информацию о проверке прав
type PermissionCheckResult struct {
	Allowed     bool   `json:"allowed"`
	Reason      string `json:"reason"`
	RequiredLvl int    `json:"required_lvl"`
	CurrentLvl  int    `json:"current_lvl"`
	IsExpired   bool   `json:"is_expired"`
	DaysLeft    int    `json:"days_left"`
	PlanName    string `json:"plan_name"`
	Permission  string `json:"permission"`
}

// CheckPermission проверяет доступ на основе тарифа компании и разрешения
func (s *Service) CheckPermission(
	ctx context.Context,
	tenantPool *pgxpool.Pool,
	companyID string,
	accountID string,
	permission string,
) (bool, error) {
	result, err := s.CheckPermissionDetailed(ctx, tenantPool, companyID, accountID, permission)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// CheckPermissionDetailed возвращает детальную информацию о проверке прав
func (s *Service) CheckPermissionDetailed(
	ctx context.Context,
	tenantPool *pgxpool.Pool,
	companyID string,
	accountID string,
	permission string,
) (*PermissionCheckResult, error) {
	// 1. Получаем текущий план компании
	companyPlan, err := s.companiesService.GetCompanyPlan(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company plan: %w", err)
	}

	// 2. Получаем требуемый lvl для разрешения
	requiredLvl := config.GetPermissionLvl(permission)
	currentLvl := companyPlan.CurrentPlan.Lvl

	result := &PermissionCheckResult{
		Permission:  permission,
		RequiredLvl: requiredLvl,
		CurrentLvl:  currentLvl,
		IsExpired:   companyPlan.DaysLeft == 0,
		DaysLeft:    companyPlan.DaysLeft,
		PlanName:    companyPlan.CurrentPlan.Name,
	}

	// 3. Проверяем по lvl
	if currentLvl > requiredLvl {
		result.Allowed = false
		result.Reason = fmt.Sprintf(
			"Tariff level too low: required level %d (permission requires tariff level %d or higher), current plan level %d (plan: %s)",
			requiredLvl, requiredLvl, currentLvl, companyPlan.CurrentPlan.Name,
		)
		return result, nil
	}

	// 4. Собираем разрешения пользователя
	userPermissions, err := s.getUserPermissions(ctx, tenantPool, companyID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 5. Проверяем наличие разрешения у пользователя
	if !userPermissions[permission] {
		result.Allowed = false
		result.Reason = fmt.Sprintf(
			"User does not have permission %s (role/permissions/position settings)",
			permission,
		)
		return result, nil
	}

	// 6. Проверяем истекший тариф
	if result.IsExpired {
		if config.IsExpiredAllowed(permission) {
			result.Allowed = true
			result.Reason = fmt.Sprintf(
				"Expired tariff allowed: permission %s is in expired allowed list, plan: %s, days left: %d, user has permission",
				permission, companyPlan.CurrentPlan.Name, companyPlan.DaysLeft,
			)
			return result, nil
		}

		result.Allowed = false
		result.Reason = fmt.Sprintf(
			"Tariff expired and permission not allowed: permission %s is not in expired allowed list, plan: %s, days left: %d",
			permission, companyPlan.CurrentPlan.Name, companyPlan.DaysLeft,
		)
		return result, nil
	}

	// 7. Доступ разрешен
	result.Allowed = true
	result.Reason = fmt.Sprintf(
		"Access granted: required level %d <= current level %d, tariff active (days left: %d), plan: %s, user has permission",
		requiredLvl, currentLvl, companyPlan.DaysLeft, companyPlan.CurrentPlan.Name,
	)
	return result, nil
}

// getUserPermissions собирает все разрешения пользователя в компании
func (s *Service) getUserPermissions(
	ctx context.Context,
	tenantPool *pgxpool.Pool,
	companyID string,
	accountID string,
) (map[string]bool, error) {
	// 1. Получаем роль аккаунта в публичной схеме
	roleCode, err := s.getAccountRole(ctx, companyID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account role: %w", err)
	}

	var basePermissions map[string]bool

	// 2. Если роль не owner - берем только гостевые разрешения
	if roleCode != companies.RoleOwner {
		basePermissions = config.GetGuestPermissions()
	} else {
		// 3. Если роль owner - берем разрешения на основе тарифа
		basePermissions, err = s.getTariffPermissions(ctx, companyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tariff permissions: %w", err)
		}
	}

	// 4. Получаем разрешения из должностей сотрудника (через тенантский пул)
	positionPermissions, err := s.getPositionPermissions(ctx, tenantPool, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position permissions: %w", err)
	}

	// 5. Объединяем базовые разрешения с должностными
	merged := make(map[string]bool)
	for p := range basePermissions {
		merged[p] = true
	}
	for _, p := range positionPermissions {
		merged[p] = true
	}

	// 6. Получаем настройки аккаунта (increase/reduce)
	increasePerms, reducePerms, err := s.getAccountSettings(ctx, tenantPool, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account settings: %w", err)
	}

	// 7. Добавляем increase разрешения
	for _, p := range increasePerms {
		merged[p] = true
	}

	// 8. Удаляем reduce разрешения
	for _, p := range reducePerms {
		delete(merged, p)
	}

	return merged, nil
}

// getAccountRole возвращает роль аккаунта в компании
func (s *Service) getAccountRole(ctx context.Context, companyID, accountID string) (string, error) {
	query := `
		SELECT role_code
		FROM company_accounts
		WHERE company_id = $1 AND account_id = $2
	`

	var roleCode string
	err := s.companiesService.GetPool().QueryRow(ctx, query, companyID, accountID).Scan(&roleCode)
	if err != nil {
		return "", fmt.Errorf("failed to get account role: %w", err)
	}

	return roleCode, nil
}

// getTariffPermissions возвращает разрешения на основе тарифа компании
func (s *Service) getTariffPermissions(ctx context.Context, companyID string) (map[string]bool, error) {
	companyPlan, err := s.companiesService.GetCompanyPlan(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company plan: %w", err)
	}

	// Получаем все разрешения, которые подходят по lvl
	allPermissions := config.GetAllPermissions()
	result := make(map[string]bool)

	for _, perm := range allPermissions {
		requiredLvl := config.GetPermissionLvl(perm)
		if companyPlan.CurrentPlan.Lvl <= requiredLvl {
			result[perm] = true
		}
	}

	return result, nil
}

// getPositionPermissions возвращает разрешения из должностей сотрудника
func (s *Service) getPositionPermissions(
	ctx context.Context,
	tenantPool *pgxpool.Pool,
	accountID string,
) ([]string, error) {
	// 1. Находим employee по account_id
	var employeeID string
	employeeQuery := `
		SELECT employee_id
		FROM employee_account
		WHERE account_id = $1
	`
	err := tenantPool.QueryRow(ctx, employeeQuery, accountID).Scan(&employeeID)
	if err != nil {
		// Если сотрудник не найден - возвращаем пустой массив
		return []string{}, nil
	}

	// 2. Получаем все должности сотрудника
	positionsQuery := `
		SELECT p.permissions
		FROM employee_position ep
		INNER JOIN employees_positions p ON ep.position_id = p.id
		WHERE ep.employee_id = $1
	`

	rows, err := tenantPool.Query(ctx, positionsQuery, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var allPermissions []string
	for rows.Next() {
		var permissionsJSON []byte
		if err := rows.Scan(&permissionsJSON); err != nil {
			return nil, fmt.Errorf("failed to scan permissions: %w", err)
		}

		var perms []string
		if len(permissionsJSON) > 0 {
			if err := json.Unmarshal(permissionsJSON, &perms); err != nil {
				return nil, fmt.Errorf("failed to parse permissions: %w", err)
			}
		}
		allPermissions = append(allPermissions, perms...)
	}

	return allPermissions, nil
}

// getAccountSettings возвращает настройки аккаунта (increase и reduce разрешения)
func (s *Service) getAccountSettings(
	ctx context.Context,
	tenantPool *pgxpool.Pool,
	accountID string,
) (increase []string, reduce []string, err error) {
	query := `
		SELECT increase_permissions, reduce_permissions
		FROM accounts_settings
		WHERE account_id = $1
	`

	var increaseJSON, reduceJSON []byte
	err = tenantPool.QueryRow(ctx, query, accountID).Scan(&increaseJSON, &reduceJSON)
	if err != nil {
		// Если записи нет - возвращаем пустые массивы
		return []string{}, []string{}, nil
	}

	if len(increaseJSON) > 0 {
		if err := json.Unmarshal(increaseJSON, &increase); err != nil {
			return nil, nil, fmt.Errorf("failed to parse increase permissions: %w", err)
		}
	}

	if len(reduceJSON) > 0 {
		if err := json.Unmarshal(reduceJSON, &reduce); err != nil {
			return nil, nil, fmt.Errorf("failed to parse reduce permissions: %w", err)
		}
	}

	return increase, reduce, nil
}
