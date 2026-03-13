package accounts

import (
	"encoding/json"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ----------
// FINGERPRINTS
// ----------

// CreateFingerprint создает новый фингерпринт
func (h *Handlers) CreateFingerprint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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

// GetFingerprints возвращает список фингерпринтов текущего пользователя
func (h *Handlers) GetFingerprints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Парсим параметры запроса
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

// Вспомогательная функция
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

// LoginWithFingerprint вход по фингерпринту
func (h *Handlers) LoginWithFingerprint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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
		core.SendUnauthorized(w, err.Error())
		return
	}

	response := FingerprintLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         account,
	}

	core.SendSuccess(w, response, "Login successful")
}

// RevokeFingerprint отзывает фингерпринт
func (h *Handlers) RevokeFingerprint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Получаем ID из URL
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
