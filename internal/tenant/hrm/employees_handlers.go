package hrm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"
)

// ---------
// EMPLOYEES
// ---------

// -------
// Цепочка действий с участием
// глобального реестра акккаунтов
// в идеале это дело потом вынести из модуля
// -------
func (h *Handlers) RemoveEmployeeAccount(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем companyID из контекста
	companyID, ok := core.GetCompanyIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusBadRequest, "Company context not found.")
		return
	}

	// Получаем accountID из URL
	targetAccountID := r.PathValue("accountId")
	if targetAccountID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Account ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Account ID is required.")
		return
	}

	// Удаляем связь
	err := h.repository.RemoveEmployeeAccount(r.Context(), companyID, targetAccountID)
	if err != nil {
		// Проверяем тип ошибки
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "cannot remove owner"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("account_id", targetAccountID),
			)
			core.SendValidationError(w, "Cannot remove owner from company")
		case strings.Contains(errorMsg, "member not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Account not found in company"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("account_id", targetAccountID),
			)
			core.SendNotFound(w, "Account not found in company")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to remove account: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("account_id", targetAccountID),
		logs.WithMetadata("action", "remove_account"),
	)

	core.SendSuccess(w, map[string]interface{}{
		"account_id": targetAccountID,
		"removed":    true,
	}, "Account removed successfully.")
}

func (h *Handlers) GetEmployee(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// Получаем сотрудника
	employee, err := h.repository.GetEmployeeByID(r.Context(), employeeId)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("employee_id", employeeId),
		)
		core.SendNotFound(w, "Employee not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("employee_id", employeeId),
	)

	core.SendSuccess(w, employee, "Employee retrieved successfully.")
}

func (h *Handlers) GetEmployees(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)
	search := r.URL.Query().Get("search")

	employees, total, err := h.repository.GetEmployees(
		r.Context(),
		pagination.Offset,
		pagination.Limit,
		search,
	)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get employees: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"search": search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(employees)),
	)

	response := map[string]interface{}{
		"employees": employees,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Employees retrieved successfully.")
}

// деактивация
func (h *Handlers) DeactivateEmployee(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// деактивируем
	ok, err := h.repository.DeactivateEmployee(r.Context(), employeeId)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("employee_id", employeeId),
		)
		core.SendInternalError(w, "Failed to deactivate employee.")
		return
	}

	if !ok {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("employee_id", employeeId),
		)
		core.SendNotFound(w, "Employee not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("employee_id", employeeId),
		logs.WithMetadata("action", "deactivate"),
	)

	core.SendSuccess(w, map[string]interface{}{
		"employee_id": employeeId,
		"deactivated": true,
	}, "Employee deactivated successfully.")
}

// активация
func (h *Handlers) ActivateEmployee(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// восстанавливаем
	ok, err := h.repository.ActivateEmployee(r.Context(), employeeId)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("employee_id", employeeId),
		)
		core.SendInternalError(w, "Failed to activate employee.")
		return
	}

	if !ok {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("employee_id", employeeId),
		)
		core.SendNotFound(w, "Employee not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("employee_id", employeeId),
		logs.WithMetadata("action", "activate"),
	)

	core.SendSuccess(w, map[string]interface{}{
		"employee_id": employeeId,
		"activated":   true,
	}, "Employee activated successfully.")
}

func (h *Handlers) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Обновляем сотрудника
	employee, err := h.repository.UpdateEmployee(r.Context(), employeeId, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "employee not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Employee not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("employee_id", employeeId),
			)
			core.SendNotFound(w, "Employee not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update employee: %s", err.Error()))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("employee_id", employeeId),
	)

	core.SendSuccess(w, employee, "Employee updated successfully.")
}

func (h *Handlers) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим тело запроса
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация
	if req.FirstName == "" || len(strings.TrimSpace(req.FirstName)) < 2 {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "First name is required and must be at least 2 characters"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "First name is required and must be at least 2 characters.")
		return
	}

	// Создаем сотрудника
	employee, err := h.repository.CreateEmployee(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create employee: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("employee_id", employee.ID),
		logs.WithMetadata("first_name", req.FirstName),
		logs.WithMetadata("last_name", req.LastName),
	)

	core.SendSuccess(w, employee, "Employee created successfully.")
}

// -------
// привязка с проверкой
// существования аккаунта в глоб.реестре
// -------
func (h *Handlers) LinkAccountEmployee(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем companyID из контекста
	companyID, ok := core.GetCompanyIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusBadRequest, "Company context not found.")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// парсим тело запроса
	var req LinkAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// проверяем принадлежность аккаунта к компании
	ok, err := h.repository.companiesService.CheckCompanyMembership(r.Context(), companyID, req.AccountId)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("employee_id", employeeId),
			logs.WithMetadata("account_id", req.AccountId),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to link account: %s", err.Error()))
		return
	}
	if !ok {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "The account does not belong to the company"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("employee_id", employeeId),
			logs.WithMetadata("account_id", req.AccountId),
		)
		core.SendError(w, http.StatusBadRequest, "The account does not belong to the company.")
		return
	}

	// ебашим
	employee, err := h.repository.LinkAccount(r.Context(), employeeId, req.AccountId)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "employee not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Employee not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("employee_id", employeeId),
			)
			core.SendNotFound(w, "Employee not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to link account: %s", err.Error()))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("employee_id", employeeId),
		logs.WithMetadata("account_id", req.AccountId),
		logs.WithMetadata("action", "link_account"),
	)

	core.SendSuccess(w, employee, "Account linked successfully.")
}

// отвязка аккаунта
func (h *Handlers) UnlinkAccountEmployee(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// ебашим
	employee, err := h.repository.UnlinkAccount(r.Context(), employeeId)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "employee not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Employee not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("employee_id", employeeId),
			)
			core.SendNotFound(w, "Employee not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to unlink account: %s", err.Error()))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_EMPLOYEES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("employee_id", employeeId),
		logs.WithMetadata("action", "unlink_account"),
	)

	core.SendSuccess(w, employee, "Account unlinked successfully.")
}
