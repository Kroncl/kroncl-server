package accounts

import (
	"context"
	"encoding/json"
	"kroncl-server/internal/config"
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

	refreshMaxAge := int(h.service.jwtService.GetRefreshDuration().Seconds())

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     config.AUTH_REFRESH_PATH,
		MaxAge:   refreshMaxAge,
	})

	data := map[string]interface{}{
		"user_id":      account.ID,
		"access_token": accessToken,
		"email_sent":   true,
	}

	core.SendCreated(w, data, "Registration successful")
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

	refreshMaxAge := int(h.service.jwtService.GetRefreshDuration().Seconds())

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     config.AUTH_REFRESH_PATH,
		MaxAge:   refreshMaxAge,
	})

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
		"access_token": accessToken,
		"user":         account,
	}

	core.SendSuccess(w, data, "Login successful")
}

func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		core.SendUnauthorized(w, "Refresh token not found")
		return
	}

	accessToken, newRefreshToken, err := h.service.RefreshTokens(r.Context(), cookie.Value)
	if err != nil {
		core.SendUnauthorized(w, err.Error())
		return
	}

	refreshMaxAge := int(h.service.jwtService.GetRefreshDuration().Seconds())

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     config.AUTH_REFRESH_PATH,
		MaxAge:   refreshMaxAge,
	})

	data := map[string]interface{}{
		"access_token": accessToken,
	}

	core.SendSuccess(w, data, "Tokens refreshed")
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     config.AUTH_REFRESH_PATH,
		MaxAge:   -1,
	})

	core.SendSuccess(w, nil, "Logged out")
}
