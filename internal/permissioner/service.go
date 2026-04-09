package permissioner

import (
	"context"
	"fmt"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
)

type Service struct {
	companiesService *companies.Service
}

func NewService(companiesService *companies.Service) *Service {
	return &Service{
		companiesService: companiesService,
	}
}

// CheckPermission проверяет доступ на основе тарифа компании и разрешения
func (s *Service) CheckPermission(
	ctx context.Context,
	companyID string,
	permission string,
) (bool, error) {
	// 1. Получаем текущий план компании
	companyPlan, err := s.companiesService.GetCompanyPlan(ctx, companyID)
	if err != nil {
		return false, fmt.Errorf("failed to get company plan: %w", err)
	}

	// 2. Получаем требуемый lvl для разрешения
	requiredLvl := config.GetPermissionLvl(permission)
	currentLvl := companyPlan.CurrentPlan.Lvl

	// 3. Проверяем по lvl
	// Чем меньше lvl, тем больше прав. Разрешено, если requiredLvl >= currentLvl
	if requiredLvl > currentLvl {
		return false, nil
	}

	// 4. Проверяем истекший тариф
	isExpired := companyPlan.DaysLeft == 0

	if isExpired {
		// Разрешаем только если разрешение в белом списке expired allowed
		if !config.IsExpiredAllowed(permission) {
			return false, nil
		}
	}

	return true, nil
}
