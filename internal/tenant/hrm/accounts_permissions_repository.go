package hrm

import (
	"context"
	"fmt"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
)

// getAccountRole возвращает роль аккаунта в компании (использует публичный пул из companiesService)
func (r *Repository) getAccountRole(ctx context.Context, companyID, accountID string) (string, error) {
	// Используем публичный пул из companiesService
	publicPool := r.companiesService.GetPool()

	query := `
		SELECT role_code
		FROM company_accounts
		WHERE company_id = $1 AND account_id = $2
	`

	var roleCode string
	err := publicPool.QueryRow(ctx, query, companyID, accountID).Scan(&roleCode)
	if err != nil {
		return "", fmt.Errorf("failed to get account role: %w", err)
	}

	return roleCode, nil
}

// getTariffPermissions возвращает разрешения на основе тарифа компании
func (r *Repository) getTariffPermissions(ctx context.Context, companyID string) (map[string]bool, error) {
	companyPlan, err := r.companiesService.GetCompanyPlan(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company plan: %w", err)
	}

	currentLvl := companyPlan.CurrentPlan.Lvl

	// Получаем все разрешения, которые подходят по lvl
	allPermissions := config.GetAllPermissions()
	result := make(map[string]bool)

	for _, perm := range allPermissions {
		requiredLvl := config.GetPermissionLvl(perm)
		if currentLvl <= requiredLvl {
			result[perm] = true
		}
	}

	return result, nil
}

// GetAccountPermissions возвращает все разрешения аккаунта в компании
func (r *Repository) GetAccountPermissions(
	ctx context.Context,
	companyID string,
	accountID string,
) (map[string]bool, error) {
	// 1. Получаем роль аккаунта в компании (публичный пул)
	roleCode, err := r.getAccountRole(ctx, companyID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account role: %w", err)
	}

	var basePermissions map[string]bool

	// 2. Если роль не owner - берем только гостевые разрешения
	if roleCode != companies.RoleOwner {
		basePermissions = config.GetGuestPermissions()
	} else {
		// 3. Если роль owner - берем разрешения на основе тарифа
		basePermissions, err = r.getTariffPermissions(ctx, companyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tariff permissions: %w", err)
		}
	}

	// 4. Получаем разрешения из должностей сотрудника (tenant пул)
	positionPermissions, err := r.getPositionPermissions(ctx, accountID)
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

	// 6. Получаем настройки аккаунта (tenant пул)
	settings, err := r.GetAccountSettings(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account settings: %w", err)
	}

	// 7. Добавляем increase разрешения
	for _, p := range settings.IncreasePermissions {
		merged[p] = true
	}

	// 8. Удаляем reduce разрешения
	for _, p := range settings.ReducePermissions {
		delete(merged, p)
	}

	return merged, nil
}
