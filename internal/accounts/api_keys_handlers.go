package accounts

import (
	"encoding/json"
	"kroncl-server/internal/core"
	"log"
	"net/http"
	"strconv"
)

func (h *Handlers) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateApiKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	if req.Name == "" {
		core.SendValidationError(w, "Name is required")
		return
	}

	key, err := h.service.CreateApiKey(r.Context(), accountID, req)
	if err != nil {
		log.Printf("❌ Failed to create api key for %s: %v", accountID, err)
		core.SendValidationError(w, err.Error())
		return
	}

	log.Printf("✅ API key created: %s for account %s", key.ID, accountID)

	core.SendCreated(w, key, "API key created")
}

func (h *Handlers) GetApiKey(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	keyID := r.PathValue("keyId")
	if keyID == "" {
		core.SendValidationError(w, "Key ID is required")
		return
	}

	key, err := h.service.GetApiKey(r.Context(), accountID, keyID)
	if err != nil {
		log.Printf("❌ Failed to get api key %s: %v", keyID, err)
		core.SendNotFound(w, "API key not found")
		return
	}

	core.SendSuccess(w, key, "API key retrieved")
}

func (h *Handlers) GetApiKeys(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	var search *string
	if s := r.URL.Query().Get("search"); s != "" {
		search = &s
	}

	var status *string
	if st := r.URL.Query().Get("status"); st != "" {
		status = &st
	}

	req := ApiKeyListRequest{
		Page:   page,
		Limit:  limit,
		Search: search,
		Status: status,
	}

	response, err := h.service.GetApiKeys(r.Context(), accountID, req)
	if err != nil {
		log.Printf("❌ Failed to list api keys for %s: %v", accountID, err)
		core.SendInternalError(w, "Failed to list api keys")
		return
	}

	core.SendSuccess(w, response, "API keys retrieved")
}

func (h *Handlers) RevokeApiKey(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	keyID := r.PathValue("keyId")
	if keyID == "" {
		core.SendValidationError(w, "Key ID is required")
		return
	}

	err := h.service.RevokeApiKey(r.Context(), accountID, keyID)
	if err != nil {
		log.Printf("❌ Failed to revoke api key %s: %v", keyID, err)
		core.SendNotFound(w, err.Error())
		return
	}

	log.Printf("✅ API key revoked: %s by account %s", keyID, accountID)

	core.SendSuccess(w, nil, "API key revoked")
}
