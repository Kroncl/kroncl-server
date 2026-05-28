package billing

import (
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"net/http"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) GetBillingMode(w http.ResponseWriter, r *http.Request) {
	mode := config.GetBillingMode()
	core.SendSuccess(w, map[string]string{"mode": mode}, "Billing mode retrieved successfully")
}
