package accounts

import (
	"context"
	"encoding/json"
	"kroncl-server/internal/core"
	"kroncl-server/internal/mailer"
	"kroncl-server/utils"
	"net/http"
	"time"
)

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	account, accessToken, refreshToken, err := h.service.Create(
		r.Context(),
		req.Email,
		req.Name,
		req.Password,
	)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	data := map[string]interface{}{
		"user_id":       account.ID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"email_sent":    true,
	}

	core.SendCreated(w, data, "Registration successful. Please check your email to confirm your account.")
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	account, accessToken, refreshToken, err := h.service.Authenticate(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		core.SendUnauthorized(w, err.Error())
		return
	}

	go func() {
		data := &mailer.LoginNotificationData{
			UserEmail: account.Email,
			UserName:  account.Name,
			IPAddress: utils.GetClientIP(r),
			LoginTime: time.Now(),
		}
		h.service.mailer.SendLoginNotification(context.Background(), data)
	}()

	data := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          account,
	}

	core.SendSuccess(w, data, "Login successful")
}

func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	if req.RefreshToken == "" {
		core.SendValidationError(w, "Refresh token is required")
		return
	}

	accessToken, refreshToken, err := h.service.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		core.SendUnauthorized(w, err.Error())
		return
	}

	data := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	core.SendSuccess(w, data, "Tokens refreshed successfully")
}
