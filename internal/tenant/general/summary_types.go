package tenantgeneral

import "kroncl-server/internal/currency"

type CompanySummary struct {
	// Сотрудники
	EmployeesTotal  int `json:"employees_total"`
	EmployeesActive int `json:"employees_active"`

	// Сделки
	DealsTotal int `json:"deals_total"`

	// Каталог
	CategoriesTotal int `json:"categories_total"`
	UnitsTotal      int `json:"units_total"`

	// Клиенты
	ClientsTotal  int `json:"clients_total"`
	ClientsActive int `json:"clients_active"`

	// Баланс организации (сконвертированный в targetCurrency)
	BalanceIncome   float64            `json:"balance_income"`
	BalanceExpense  float64            `json:"balance_expense"`
	BalanceTotal    float64            `json:"balance_total"`
	BalanceCurrency *currency.Currency `json:"currency"`
	BalanceCount    int64              `json:"balance_count"`
}
