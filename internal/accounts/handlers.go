package accounts

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/core"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Handlers содержит HTTP хендлеры для аккаунтов
type Handlers struct {
	service *Service
}

// NewHandlers создает новый экземпляр хендлеров
func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

// Register обрабатывает запрос на регистрацию
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Создаем аккаунт и получаем токены
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

	// Формируем данные для ответа
	data := map[string]interface{}{
		"user_id":       account.ID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"email_sent":    true,
	}

	// Отправляем ответ
	core.SendCreated(w, data, "Registration successful. Please check your email to confirm your account.")
}

// Login обрабатывает запрос на вход
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Аутентификация
	account, accessToken, refreshToken, err := h.service.Authenticate(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		core.SendUnauthorized(w, err.Error())
		return
	}

	// Формируем данные
	data := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          account,
	}

	core.SendSuccess(w, data, "Login successful")
}

// ConfirmEmail подтверждает email
func (h *Handlers) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
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

	// Проверяем, что user_id в запросе совпадает с токеном
	if req.UserID != claims.UserID {
		core.SendUnauthorized(w, "User ID mismatch")
		return
	}

	// Подтверждаем email
	err := h.service.ConfirmEmail(r.Context(), req.UserID, req.Code)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	// Ответ с пустыми данными
	core.SendSuccess(w, map[string]interface{}{}, "Email confirmed successfully")
}

// Повторная отправка кода подтверждения
func (h *Handlers) ResendConfirmationCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Повторяем отправку кода
	err := h.service.ResendConfirmationCode(r.Context(), claims.UserID)
	if err != nil {
		// Просто отправляем ошибку как есть, middleware обработает
		core.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Ответ с пустыми данными
	core.SendSuccess(w, map[string]interface{}{}, "Confirmation code has been resent to your email")
}

// GetProfile получает профиль пользователя
func (h *Handlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	log.Printf("user id: %s", claims.UserID)

	// Получаем аккаунт из БД
	account, err := h.service.GetByID(r.Context(), claims.UserID)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("User not found: %s", err.Error()))
		return
	}

	// Отправляем профиль
	core.SendSuccess(w, account, "Profile retrieved successfully")
}

func (h *Handlers) CheckEmailUnique(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем email из query параметров
	email := r.URL.Query().Get("email")
	if email == "" {
		core.SendValidationError(w, "email parameter is required")
		return
	}

	// Валидация email
	if !isValidEmail(email) {
		core.SendValidationError(w, "Invalid email format")
		return
	}

	// Проверяем уникальность
	unique, err := h.service.checkEmailUnique(r.Context(), email)
	if err != nil {
		core.SendInternalError(w, err.Error())
		return
	}

	if !unique {
		core.SendValidationError(w, "The mail is not unique")
		return
	}

	core.SendSuccess(w, map[string]interface{}{}, "The mail is unique")
}

// Вспомогательная функция для валидации email
func isValidEmail(email string) bool {
	// Простая проверка, можно использовать regex или validator
	if len(email) > 254 {
		return false
	}

	// Проверяем наличие @ и точки после нее
	at := strings.LastIndex(email, "@")
	if at < 1 || at > len(email)-4 {
		return false
	}

	dot := strings.LastIndex(email[at:], ".")
	if dot < 2 || dot > len(email[at:])-3 {
		return false
	}

	return true
}

// Refresh обновляет токены по refresh токену
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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

	// Обновляем токены
	accessToken, refreshToken, err := h.service.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		core.SendUnauthorized(w, err.Error())
		return
	}

	// Формируем ответ
	data := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	core.SendSuccess(w, data, "Tokens refreshed successfully")
}

// обновление данных пользователя (avatar/name)
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Incorrect account data.")
		return
	}

	// Получаем аккаунт из БД
	account, err := h.service.UpdateById(r.Context(), claims.UserID, &req)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("User update error: %s", err.Error()))
		return
	}

	// Отправляем профиль
	core.SendSuccess(w, account, "Profile updated successfully")
}

// GetPublicAccounts возвращает список аккаунтов с пагинацией и поиском
func (h *Handlers) GetPublicAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Извлекаем параметры поиска
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	// Получаем параметры пагинации
	paginationParams := core.GetDefaultPaginationParams(r)

	var accounts []AccountPublic
	var pagination core.Pagination

	accounts, pagination, err := h.service.GetPublicAccounts(
		r.Context(),
		search,
		paginationParams,
	)

	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Error receiving accounts: %s", err.Error()))
		return
	}

	// Формируем ответ
	response := map[string]interface{}{
		"accounts":   accounts,
		"pagination": pagination,
	}

	core.SendSuccess(w, response, "Accounts retrieved successfully")
}

// GetAccountInvitations возвращает приглашения для текущего пользователя
func (h *Handlers) GetAccountInvitations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Получаем параметры запроса
	query := r.URL.Query()

	// Фильтр по статусу
	status := query.Get("status")

	// Пагинация
	paginationParams := core.GetDefaultPaginationParams(r)

	// Формируем запрос
	req := companies.GetInvitationsByEmailRequest{
		Status:           status,
		PaginationParams: paginationParams,
	}

	// Получаем приглашения
	response, err := h.service.GetAccountInvitations(r.Context(), claims.UserID, req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get invitations: %v", err))
		return
	}

	// Отправляем ответ
	core.SendSuccess(w, response, "Invitations retrieved successfully")
}

// AcceptAccountInvitation принимает приглашение
func (h *Handlers) AcceptAccountInvitation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Извлекаем ID приглашения из URL параметра
	invitationID := chi.URLParam(r, "invitationId")

	if invitationID == "" {
		core.SendValidationError(w, "Invitation ID is required")
		return
	}

	// Принимаем приглашение
	invitation, err := h.service.AcceptInvitation(r.Context(), claims.UserID, invitationID)
	if err != nil {
		// Определяем тип ошибки для соответствующего HTTP статуса
		switch {
		case strings.Contains(err.Error(), "account must be confirmed"):
			core.SendValidationError(w, "You must confirm your email before accepting invitations")
		case strings.Contains(err.Error(), "invitation does not belong"):
			core.SendUnauthorized(w, "This invitation does not belong to you")
		case strings.Contains(err.Error(), "invitation not found"):
			core.SendNotFound(w, "Invitation not found")
		case strings.Contains(err.Error(), "invitation is not in waiting status"):
			core.SendValidationError(w, "This invitation is no longer valid")
		case strings.Contains(err.Error(), "is already a member"):
			core.SendValidationError(w, "You are already a member of this company")
		case strings.Contains(err.Error(), "user is already a member"):
			core.SendValidationError(w, "You are already a member of this company")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to accept invitation: %v", err))
		}
		return
	}

	core.SendSuccess(w, invitation, "Invitation accepted successfully")
}

// RejectAccountInvitation отклоняет приглашение
func (h *Handlers) RejectAccountInvitation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Извлекаем ID приглашения из URL параметра
	invitationID := chi.URLParam(r, "invitationId")

	if invitationID == "" {
		core.SendValidationError(w, "Invitation ID is required")
		return
	}

	// Отклоняем приглашение
	invitation, err := h.service.RejectInvitation(r.Context(), claims.UserID, invitationID)
	if err != nil {
		// Определяем тип ошибки
		switch {
		case strings.Contains(err.Error(), "account must be confirmed"):
			core.SendValidationError(w, "You must confirm your email before rejecting invitations")
		case strings.Contains(err.Error(), "invitation does not belong"):
			core.SendUnauthorized(w, "This invitation does not belong to you")
		case strings.Contains(err.Error(), "invitation not found"):
			core.SendNotFound(w, "Invitation not found")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to reject invitation: %v", err))
		}
		return
	}

	core.SendSuccess(w, invitation, "Invitation rejected successfully")
}

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
