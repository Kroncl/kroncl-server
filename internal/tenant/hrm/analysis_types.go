package hrm

type GroupBy string

const (
	GroupByDay   GroupBy = "day"
	GroupByMonth GroupBy = "month"
	GroupByYear  GroupBy = "year"
)

type GroupedStats struct {
	GroupKey       string `json:"group_key"`       // date in format: YYYY-MM-DD, YYYY-MM, YYYY
	GroupName      string `json:"group_name"`      // formatted date
	EmployeesCount int64  `json:"employees_count"` // количество сотрудников
	ActiveCount    int64  `json:"active_count"`    // количество активных сотрудников
	InactiveCount  int64  `json:"inactive_count"`  // количество неактивных сотрудников
}

// EmployeesSummary represents summary statistics for a period
type EmployeesSummary struct {
	TotalPositions    int64 `json:"total_positions"`    // общее количество должностей
	TotalEmployees    int64 `json:"total_employees"`    // общее количество сотрудников
	ActiveEmployees   int64 `json:"active_employees"`   // количество активных сотрудников
	InactiveEmployees int64 `json:"inactive_employees"` // количество неактивных сотрудников
	NewEmployees      int64 `json:"new_employees"`      // новые сотрудники за период
}
