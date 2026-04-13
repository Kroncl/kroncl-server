package accounts

import (
	"encoding/json"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"net/http"
)

func (h *Handlers) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	if req.UserID != claims.UserID {
		core.SendUnauthorized(w, "User ID mismatch")
		return
	}

	err := h.service.ConfirmEmail(r.Context(), req.UserID, req.Code)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	core.SendSuccess(w, map[string]interface{}{}, "Email confirmed successfully")
}

func (h *Handlers) ResendConfirmationCode(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	err := h.service.ResendConfirmationCode(r.Context(), claims.UserID)
	if err != nil {
		core.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	core.SendSuccess(w, map[string]interface{}{}, "Confirmation code has been resent to your email")
}
