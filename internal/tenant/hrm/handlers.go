package hrm

import (
	"encoding/json"
	"kroncl-server/internal/core"
	"net/http"
)

type Handlers struct {
	repository *Repository
}

func NewHandlers(repository *Repository) *Handlers {
	return &Handlers{repository: repository}
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

	// Получаем параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Получаем сотрудников
	employees, total, err := h.repository.GetEmployees(
		r.Context(),
		pagination.Offset,
		pagination.Limit,
	)
	if err != nil {
		core.SendError(w, http.StatusInternalServerError, "Failed to get employees.")
		return
	}

	// Создаем ответ
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
		core.SendError(w, http.StatusInternalServerError, "Failed to update employee.")
		return
	}

	core.SendSuccess(w, employee, "Employee updated successfully.")
}
