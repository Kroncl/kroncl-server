package companies

import (
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"net/http"
)

// GetPermissions возвращает все разрешения с критичностью и требуемым уровнем тарифа
func (h *Handlers) GetPermissions(w http.ResponseWriter, r *http.Request) {
	// Получаем все разрешения с критичностью
	criticalityMap := config.GetCriticalityLevels()

	// Формируем ответ
	type PermissionItem struct {
		Code        string `json:"code"`
		Criticality int    `json:"criticality"`
		RequiredLvl int    `json:"lvl"`
	}

	items := make([]PermissionItem, 0, len(criticalityMap))
	for code, criticality := range criticalityMap {
		items = append(items, PermissionItem{
			Code:        code,
			Criticality: criticality,
			RequiredLvl: config.GetPermissionLvl(code),
		})
	}

	core.SendSuccess(w, items, "Permissions retrieved successfully")
}
