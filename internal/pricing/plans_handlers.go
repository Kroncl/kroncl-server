package pricing

import (
	"fmt"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetPlans возвращает список тарифных планов с пагинацией
func (h *Handlers) GetPlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Парсим параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Поиск
	search := r.URL.Query().Get("search")

	// Получаем планы через сервис
	plans, total, err := h.service.GetPlans(r.Context(), pagination.Page, pagination.Limit, search)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get pricing plans: %v", err))
		return
	}

	// Формируем ответ с пагинацией
	response := map[string]interface{}{
		"plans": plans,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Pricing plans retrieved successfully")
}

// GetPlanByCode возвращает тарифный план по коду
func (h *Handlers) GetPlanByCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем код плана из URL
	code := chi.URLParam(r, "code")
	if code == "" {
		core.SendValidationError(w, "Plan code is required")
		return
	}

	// Получаем план через сервис
	plan, err := h.service.GetPlanByCode(r.Context(), code)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Pricing plan with code '%s' not found", code))
		return
	}

	core.SendSuccess(w, plan, "Pricing plan retrieved successfully")
}
