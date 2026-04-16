package public

import (
	"encoding/json"
	"kroncl-server/internal/core"
	"log"
	"net/http"
)

func (h *Handlers) CreatePartner(w http.ResponseWriter, r *http.Request) {
	var req CreateIncomingPartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}
	_, err := h.service.Create(r.Context(), req)
	if err != nil {
		log.Printf("❌ Failed to create partner request for %s: %v", req.Email, err)
		core.SendValidationError(w, err.Error())
		return
	}

	core.SendSuccess(w, nil, "Request received")
}
