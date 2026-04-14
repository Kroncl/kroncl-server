package accounts

import (
	"context"
	"encoding/json"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/mailer"
	"kroncl-server/utils"
	"log"
	"net/http"
	"time"
)

func (h *Handlers) setRefreshCookie(w http.ResponseWriter, token string) {
	refreshMaxAge := int(h.service.jwtService.GetRefreshDuration().Seconds())

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   config.GetCookieSecure(),
		SameSite: config.GetCookieSameSite(),
		Path:     config.AUTH_REFRESH_PATH,
		Domain:   config.GetCookieDomain(),
		MaxAge:   refreshMaxAge,
	}

	http.SetCookie(w, cookie)

	log.Printf("🍪 Refresh cookie set: secure=%v, sameSite=%v, domain=%s, maxAge=%d",
		cookie.Secure, cookie.SameSite, cookie.Domain, cookie.MaxAge)
}

func (h *Handlers) clearRefreshCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   config.GetCookieSecure(),
		SameSite: config.GetCookieSameSite(),
		Path:     config.AUTH_REFRESH_PATH,
		Domain:   config.GetCookieDomain(),
		MaxAge:   -1,
	}

	http.SetCookie(w, cookie)

	log.Printf("🍪 Refresh cookie cleared: domain=%s", cookie.Domain)
}

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
		log.Printf("❌ Register failed for %s: %v", req.Email, err)
		core.SendValidationError(w, err.Error())
		return
	}

	h.setRefreshCookie(w, refreshToken)

	log.Printf("✅ User registered: %s (%s)", account.Email, account.ID)

	data := map[string]interface{}{
		"user_id":      account.ID,
		"access_token": accessToken,
		"expires_at":   h.service.jwtService.GetAccessDuration(),
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
		log.Printf("❌ Login failed for %s: %v", req.Email, err)
		core.SendUnauthorized(w, err.Error())
		return
	}

	h.setRefreshCookie(w, refreshToken)

	log.Printf("✅ User logged in: %s (%s) from %s", account.Email, account.ID, utils.GetClientIP(r))

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
		"expires_at":   h.service.jwtService.GetAccessDuration(),
		"user":         account,
	}

	core.SendSuccess(w, data, "Login successful")
}

func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		log.Printf("❌ Refresh failed: no cookie")
		core.SendUnauthorized(w, "Refresh token not found")
		return
	}

	accessToken, newRefreshToken, err := h.service.RefreshTokens(r.Context(), cookie.Value)
	if err != nil {
		log.Printf("❌ Refresh failed: %v", err)
		core.SendUnauthorized(w, err.Error())
		return
	}

	h.setRefreshCookie(w, newRefreshToken)

	log.Printf("🔄 Tokens refreshed")

	data := map[string]interface{}{
		"access_token": accessToken,
		"expires_at":   h.service.jwtService.GetAccessDuration(),
	}

	core.SendSuccess(w, data, "Tokens refreshed")
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearRefreshCookie(w)

	userID, _ := core.GetUserIDFromContext(r.Context())
	log.Printf("👋 User logged out: %s", userID)

	core.SendSuccess(w, nil, "Logged out")
}
