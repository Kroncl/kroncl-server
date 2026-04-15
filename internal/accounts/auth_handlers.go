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

func (h *Handlers) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	if req.Email == "" {
		core.SendValidationError(w, "Email is required")
		return
	}

	account, err := h.service.GetByEmail(r.Context(), req.Email)
	if err != nil {
		core.SendSuccess(w, nil, "If the email exists, a reset link has been sent")
		return
	}

	token, err := h.service.jwtService.GenerateResetPasswordToken(account.ID)
	if err != nil {
		log.Printf("❌ Failed to generate reset token for %s: %v", req.Email, err)
		core.SendInternalError(w, "Failed to process request")
		return
	}

	go func() {
		data := &mailer.PasswordResetData{
			UserEmail: account.Email,
			UserName:  account.Name,
			Token:     token,
		}
		h.service.mailer.SendPasswordReset(context.Background(), data)
	}()

	log.Printf("✅ Password reset requested for %s", req.Email)
	core.SendSuccess(w, nil, "If the email exists, a reset link has been sent")
}

func (h *Handlers) ValidateResetToken(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	if req.Token == "" {
		core.SendValidationError(w, "Token is required")
		return
	}

	claims, err := h.service.jwtService.ValidateResetPasswordToken(req.Token)
	if err != nil {
		core.SendValidationError(w, "Invalid or expired token")
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"account_id": claims.AccountID,
		"valid":      true,
	}, "Token is valid")
}

func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	if req.Token == "" {
		core.SendValidationError(w, "Token is required")
		return
	}

	if req.NewPassword == "" {
		core.SendValidationError(w, "New password is required")
		return
	}

	claims, err := h.service.jwtService.ValidateResetPasswordToken(req.Token)
	if err != nil {
		core.SendValidationError(w, "Invalid or expired token")
		return
	}

	err = h.service.ResetPassword(r.Context(), claims.AccountID, req.NewPassword)
	if err != nil {
		log.Printf("❌ Failed to reset password for %s: %v", claims.AccountID, err)
		core.SendValidationError(w, err.Error())
		return
	}

	go func() {
		account, err := h.service.GetByID(context.Background(), claims.AccountID)
		if err != nil {
			log.Printf("⚠️ Failed to get account for password reset success email: %v", err)
			return
		}

		data := &mailer.PasswordResetSuccessData{
			UserEmail: account.Email,
			UserName:  account.Name,
		}
		h.service.mailer.SendPasswordResetSuccess(context.Background(), data)
	}()

	log.Printf("✅ Password reset for account %s", claims.AccountID)
	core.SendSuccess(w, nil, "Password successfully reset")
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
		"expires_at":   h.service.jwtService.GetAccessExpiresAt(),
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
		"expires_at":   h.service.jwtService.GetAccessExpiresAt(),
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
		"expires_at":   h.service.jwtService.GetAccessExpiresAt(),
	}

	core.SendSuccess(w, data, "Tokens refreshed")
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearRefreshCookie(w)

	userID, _ := core.GetUserIDFromContext(r.Context())
	log.Printf("👋 User logged out: %s", userID)

	core.SendSuccess(w, nil, "Logged out")
}

// ---------
// UTILS
// ---------

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
