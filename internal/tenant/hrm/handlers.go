package hrm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/core"
	"net/http"
	"strings"
)

type Handlers struct {
	repository *Repository
}

func NewHandlers(repository *Repository) *Handlers {
	return &Handlers{repository: repository}
}

func (h *Handlers) RemoveEmployeeAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем companyID из контекста
	companyID, ok := core.GetCompanyIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusBadRequest, "Company context not found.")
		return
	}

	// Получаем accountID из URL
	accountID := r.PathValue("accountId")
	if accountID == "" {
		core.SendError(w, http.StatusBadRequest, "Account ID is required.")
		return
	}

	// Удаляем связь
	err := h.repository.RemoveEmployeeAccount(r.Context(), companyID, accountID)
	if err != nil {
		// Проверяем тип ошибки
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "cannot remove owner"):
			core.SendValidationError(w, "Cannot remove owner from company")
		case strings.Contains(errorMsg, "member not found"):
			core.SendNotFound(w, "Account not found in company")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to remove account: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"account_id": accountID,
		"removed":    true,
	}, "Account removed successfully.")
}

func (h *Handlers) GetEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// Получаем сотрудника
	employee, err := h.repository.GetEmployeeByID(r.Context(), employeeId)
	if err != nil {
		core.SendError(w, http.StatusNotFound, "Employee not found.")
		return
	}

	core.SendSuccess(w, employee, "Employee retrieved successfully.")
}

func (h *Handlers) GetEmployees(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
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
		core.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get employees: %s", err.Error()))
		return
	}

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

func (h *Handlers) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// удаляем
	ok, err := h.repository.DeleteEmployee(r.Context(), employeeId)
	if err != nil {
		core.SendError(w, http.StatusInternalServerError, "Failed to delete employee.")
		return
	}

	core.SendSuccess(w, ok, "Employee deleted successfully.")
}

func (h *Handlers) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID сотрудника из URL параметра employeeId
	employeeId := r.PathValue("employeeId")
	if employeeId == "" {
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Обновляем сотрудника
	employee, err := h.repository.UpdateEmployee(r.Context(), employeeId, req)
	if err != nil {
		core.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update employee. %s", err.Error()))
		return
	}

	core.SendSuccess(w, employee, "Employee updated successfully.")
}

func (h *Handlers) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Парсим тело запроса
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация
	if req.FirstName == "" || len(strings.TrimSpace(req.FirstName)) < 2 {
		core.SendError(w, http.StatusBadRequest, "First name is required and must be at least 2 characters.")
		return
	}

	// Создаем сотрудника
	employee, err := h.repository.CreateEmployee(r.Context(), req)
	if err != nil {
		core.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create employee: %s", err.Error()))
		return
	}

	core.SendSuccess(w, employee, "Employee created successfully.")
}
