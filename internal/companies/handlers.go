package companies

import (
	"encoding/json"
	"kroncl-server/internal/auth"
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

// обновление организации
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Отправляем ответ
	core.SendCreated(w, nil, "Company updated successful.")
}

// создание организации
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Получаем пользователя из контекста
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Создаем аккаунт и получаем токены
	data, err := h.service.Create(
		r.Context(),
		account.UserID,
		req.Slug,
		req.Name,
		req.Description,
		req.AvatarUrl,
		req.IsPublic,
	)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	// Отправляем ответ
	core.SendCreated(w, data, "Company created successful.")
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

	ok, err := h.service.checkSlugUnique(r.Context(), slug)
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
