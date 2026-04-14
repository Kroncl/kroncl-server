package accounts

import (
	"context"
	"encoding/json"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"kroncl-server/internal/mailer"
	"kroncl-server/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// ----------
// FINGERPRINTS
// ----------

func (h *Handlers) CreateFingerprint(w http.ResponseWriter, r *http.Request) {

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	var req FingerprintCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	fp, err := h.service.CreateFingerprint(r.Context(), claims.UserID, req.ExpiresIn)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	core.SendSuccess(w, fp, "Fingerprint created successfully")
}

func (h *Handlers) GetFingerprints(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	req := FingerprintListRequest{
		Page:  parseInt(r.URL.Query().Get("page"), 1),
		Limit: parseInt(r.URL.Query().Get("limit"), 20),
	}

	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = &status
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	fingerprints, err := h.service.GetAccountFingerprints(r.Context(), claims.UserID, req)
	if err != nil {
		core.SendInternalError(w, err.Error())
		return
	}

	core.SendSuccess(w, fingerprints, "Fingerprints retrieved successfully")
}

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil || val < 1 {
		return defaultValue
	}
	return val
}

func (h *Handlers) LoginWithFingerprint(w http.ResponseWriter, r *http.Request) {
	var req FingerprintLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	if req.Key == "" {
		core.SendValidationError(w, "Fingerprint key is required")
		return
	}

	accessToken, refreshToken, account, err := h.service.LoginWithFingerprint(r.Context(), req.Key)
	if err != nil {
		log.Printf("❌ Fingerprint login failed: %v", err)
		core.SendUnauthorized(w, err.Error())
		return
	}

	h.setRefreshCookie(w, refreshToken)

	log.Printf("✅ Fingerprint login: %s (%s) from %s", account.Email, account.ID, utils.GetClientIP(r))

	go func() {
		data := &mailer.LoginNotificationData{
			UserEmail: account.Email,
			UserName:  account.Name,
			IPAddress: utils.GetClientIP(r),
			LoginTime: time.Now(),
		}
		h.service.mailer.SendLoginNotification(context.Background(), data)
	}()

	response := FingerprintLoginResponse{
		AccessToken: accessToken,
		User:        account,
		ExpiresAt:   h.service.jwtService.GetAccessExpiresAt(),
	}

	core.SendSuccess(w, response, "Login successful")
}

func (h *Handlers) RevokeFingerprint(w http.ResponseWriter, r *http.Request) {

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	fingerprintID := chi.URLParam(r, "fingerprintId")
	if fingerprintID == "" {
		core.SendValidationError(w, "Fingerprint ID is required")
		return
	}

	err := h.service.RevokeFingerprint(r.Context(), claims.UserID, fingerprintID)
	if err != nil {
		if strings.Contains(err.Error(), "does not belong") {
			core.SendUnauthorized(w, err.Error())
			return
		}
		core.SendValidationError(w, err.Error())
		return
	}

	core.SendSuccess(w, map[string]interface{}{}, "Fingerprint revoked successfully")
}
