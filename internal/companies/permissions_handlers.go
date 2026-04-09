package companies

import (
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"net/http"
)

// GetPermissions возвращает все разрешения с критичностью, требуемым уровнем тарифа и флагом allowExpired
func (h *Handlers) GetPlatformPermissions(w http.ResponseWriter, r *http.Request) {
	// Получаем все разрешения с критичностью
	criticalityMap := config.GetCriticalityLevels()

	// Формируем ответ
	type PermissionItem struct {
		Code         string `json:"code"`
		Criticality  int    `json:"criticality"`
		RequiredLvl  int    `json:"lvl"`
		AllowExpired bool   `json:"allow_expired"` // доступно после истечения тарифа
	}

	items := make([]PermissionItem, 0, len(criticalityMap))
	for code, criticality := range criticalityMap {
		items = append(items, PermissionItem{
			Code:         code,
			Criticality:  criticality,
			RequiredLvl:  config.GetPermissionLvl(code),
			AllowExpired: config.IsExpiredAllowed(code),
		})
	}

	core.SendSuccess(w, items, "Permissions retrieved successfully")
}

// GetPlatformPermissions возвращает все разрешения для текущей компании
// с учетом тарифного плана компании (фильтрует по lvl)
func (h *Handlers) GetCompanyPermissions(w http.ResponseWriter, r *http.Request) {
	// Получаем ID компании из контекста
	companyID, ok := core.GetCompanyIDFromContext(r.Context())
	if !ok {
		core.SendValidationError(w, "Company ID required")
		return
	}

	// Получаем текущий план компании
	companyPlan, err := h.service.GetCompanyPlan(r.Context(), companyID)
	if err != nil {
		core.SendInternalError(w, "Failed to get company plan")
		return
	}

	currentLvl := companyPlan.CurrentPlan.Lvl

	// Получаем все разрешения с критичностью
	criticalityMap := config.GetCriticalityLevels()

	// Формируем ответ
	type PermissionItem struct {
		Code         string `json:"code"`
		Criticality  int    `json:"criticality"`
		RequiredLvl  int    `json:"lvl"`
		AllowExpired bool   `json:"allow_expired"` // доступно после истечения тарифа
		Available    bool   `json:"available"`     // доступно по текущему тарифу
	}

	items := make([]PermissionItem, 0, len(criticalityMap))
	for code, criticality := range criticalityMap {
		requiredLvl := config.GetPermissionLvl(code)

		items = append(items, PermissionItem{
			Code:         code,
			Criticality:  criticality,
			RequiredLvl:  requiredLvl,
			AllowExpired: config.IsExpiredAllowed(code),
			Available:    currentLvl <= requiredLvl,
		})
	}

	core.SendSuccess(w, items, "Platform permissions retrieved successfully")
}
