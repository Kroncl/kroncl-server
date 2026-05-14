package companies

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetCompanyVisitCard(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		core.SendValidationError(w, "Slug required.")
		return
	}

	company, err := h.service.GetCompanyVisitCard(r.Context(), slug)
	if err != nil {
		core.SendNotFound(w, "Company not found or not public")
		return
	}

	core.SendSuccess(w, company, "Company retrieved successfully.")
}

func (h *Handlers) GetCompanyMember(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	memberID := chi.URLParam(r, "accountId")
	if memberID == "" {
		core.SendValidationError(w, "Member ID required.")
		return
	}

	member, err := h.service.GetCompanyMember(r.Context(), companyID, memberID)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Company member not found: %v", err))
		return
	}

	core.SendSuccess(w, member, "Company member retrieved successfully.")
}

func (h *Handlers) GetCompanyMembers(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	req := &GetCompanyMembersRequest{
		Page:      pagination.Page,
		Limit:     pagination.Limit,
		Search:    r.URL.Query().Get("search"),
		Role:      r.URL.Query().Get("role"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	if req.Role != "" && req.Role != "all" {
		validRoles := map[string]bool{
			"owner":  true,
			"admin":  true,
			"member": true,
			"guest":  true,
		}
		if !validRoles[req.Role] {
			core.SendValidationError(w, "Invalid role. Allowed values: owner, admin, member, guest, all")
			return
		}
	}

	response, err := h.service.GetCompanyMembers(r.Context(), companyID, req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get company members: %v", err))
		return
	}

	core.SendSuccess(w, response, "Company members retrieved successfully.")
}

func (h *Handlers) GetUserCompanyById(w http.ResponseWriter, r *http.Request) {
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required.")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	data, err := h.service.GetUserCompanyById(r.Context(), account.UserID, companyID)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Company not found: %v", err))
		return
	}

	core.SendSuccess(w, data, "Company retrieved successfully.")
}

func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Incorrect company data.")
		return
	}

	updatedCompany, err := h.service.UpdateById(r.Context(), account.UserID, companyID, &req)
	if err != nil {
		core.SendValidationError(w, fmt.Sprintf("Company update error: %v", err))
		return
	}

	core.SendSuccess(w, updatedCompany, "Company updated successfully.")
}

func (h *Handlers) GetUserCompanies(w http.ResponseWriter, r *http.Request) {
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	query := r.URL.Query()

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	role := query.Get("role")
	if role == "" {
		role = "all"
	}

	search := query.Get("search")

	req := &GetUserCompaniesRequest{
		Page:   page,
		Limit:  limit,
		Role:   role,
		Search: search,
	}

	response, err := h.service.GetUserCompanies(r.Context(), account.UserID, req)
	if err != nil {
		if err.Error() == "invalid role filter. Allowed values: all, owner, admin, member, guest" {
			core.SendValidationError(w, err.Error())
		} else {
			core.SendInternalError(w, err.Error())
		}
		return
	}

	core.SendSuccess(w, response, "User companies retrieved successfully")
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	if req.Region == "" {
		req.Region = RegionRu
	}

	data, err := h.service.Create(
		r.Context(),
		account.UserID,
		req.Slug,
		req.Name,
		req.Description,
		req.AvatarUrl,
		req.IsPublic,
		req.PlanCode,
		req.Region,
		req.Promocode,
	)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	core.SendCreated(w, data, "Company created successfully.")
}

func (h *Handlers) CheckSlugUnique(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		core.SendValidationError(w, "slug parameter is required")
		return
	}

	ok, err := h.service.checkSlugUnique(r.Context(), slug)
	if err != nil {
		core.SendInternalError(w, err.Error())
		return
	}

	if !ok {
		core.SendValidationError(w, "The slug is not unique")
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"slug":   slug,
		"unique": true,
	}, "The slug is unique")
}

func (h *Handlers) Drop(w http.ResponseWriter, r *http.Request) {
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	_, err := h.service.GetUserCompanyById(r.Context(), account.UserID, companyID)
	if err != nil {
		core.SendNotFound(w, "Company not found")
		return
	}

	if err := h.service.Drop(r.Context(), companyID); err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to drop company: %v", err))
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"company_id": companyID,
		"dropped":    true,
	}, "Company dropped successfully")
}
