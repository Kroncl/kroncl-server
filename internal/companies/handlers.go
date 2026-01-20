package companies

import (
	"kroncl-server/internal/core"
	"net/http"
)

// Handlers содержит HTTP хендлеры для компаний
type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

// проверка уникальности slug компании
func (h *Handlers) CheckSlugUnique(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		core.SendValidationError(w, "slug parameter is required")
		return
	}

	ok, err := h.service.checkSlugUnique(slug)
	if err != nil {
		core.SendInternalError(w, err.Error())
		return
	}

	if !ok {
		core.SendValidationError(w, "The slug is not unique")
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"slug":   slug,
		"unique": true,
	}, "The slug is unique")
}
