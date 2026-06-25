package dadata

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/core"
	"net/http"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) FindByINN(w http.ResponseWriter, r *http.Request) {
	var req FindPartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Query == "" {
		core.SendError(w, http.StatusBadRequest, "INN is required")
		return
	}

	party, err := h.service.FindPartyByINN(r.Context(), req.Query)
	if err != nil {
		core.SendError(w, http.StatusNotFound, fmt.Sprintf("Company not found: %s", err.Error()))
		return
	}

	preview := h.service.BuildCounterpartyPreview(party)
	core.SendSuccess(w, preview, "Counterparty preview built successfully")
}

func (h *Handlers) SuggestParty(w http.ResponseWriter, r *http.Request) {
    var req FindPartyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        core.SendError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    if req.Query == "" {
        core.SendError(w, http.StatusBadRequest, "Query is required")
        return
    }

    suggestions, err := h.service.SuggestParty(r.Context(), req.Query)
    if err != nil {
        core.SendError(w, http.StatusInternalServerError, fmt.Sprintf("DaData error: %s", err.Error()))
        return
    }

    previews := make([]CounterpartyPreview, 0, len(suggestions))
    for _, s := range suggestions {
        previews = append(previews, *h.service.BuildCounterpartyPreview(&s))
    }

    core.SendSuccess(w, previews, "Suggestions retrieved successfully")
}