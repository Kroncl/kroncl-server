package admindb

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) MigrateAllTenants(w http.ResponseWriter, r *http.Request) {
	err := h.service.migrator.MigrateAllTenants(r.Context(), "up", 0)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to migrate all tenants: %v", err))
		return
	}

	core.SendSuccess(w, nil, "All tenants migrated successfully")
}

func (h *Handlers) MigrateTenant(w http.ResponseWriter, r *http.Request) {
	schemaName := chi.URLParam(r, "schemaName")
	if schemaName == "" {
		core.SendValidationError(w, "schemaName is required")
		return
	}

	var req MigrateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request body")
		return
	}

	if req.Command == "" {
		req.Command = "up"
	}

	// Валидация команды
	validCommands := map[string]bool{
		"up":      true,
		"down":    true,
		"version": true,
		"force":   true,
	}
	if !validCommands[req.Command] {
		core.SendValidationError(w, "Invalid command. Allowed: up, down, version, force")
		return
	}

	// Для force нужен steps > 0
	if req.Command == "force" && req.Steps <= 0 {
		core.SendValidationError(w, "Force command requires steps > 0")
		return
	}

	// Выполняем миграцию
	err := h.service.migrator.Run(r.Context(), schemaName, req.Command, req.Steps)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to migrate tenant: %v", err))
		return
	}

	message := fmt.Sprintf("Tenant %s migrated successfully with command '%s'", schemaName, req.Command)
	if req.Command == "down" && req.Steps > 0 {
		message = fmt.Sprintf("Tenant %s rolled back %d steps", schemaName, req.Steps)
	}
	if req.Command == "force" {
		message = fmt.Sprintf("Tenant %s forced to version %d", schemaName, req.Steps)
	}

	core.SendSuccess(w, nil, message)
}
