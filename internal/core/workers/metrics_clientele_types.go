package coreworkers

import "time"

type MetricsClienteleSnapshot struct {
	RecordedAt time.Time `json:"recorded_at"`

	// Аккаунты
	TotalAccounts     int `json:"total_accounts"`
	ConfirmedAccounts int `json:"confirmed_accounts"`
	WaitingAccounts   int `json:"waiting_accounts"`
	AdminAccounts     int `json:"admin_accounts"`

	// Статистика по типам аккаунтов
	AccountTypeOwner       int `json:"account_type_owner"`
	AccountTypeEmployee    int `json:"account_type_employee"`
	AccountTypeAdmin       int `json:"account_type_admin"`
	AccountTypeOutsourcing int `json:"account_type_outsourcing"`
	AccountTypeTech        int `json:"account_type_tech"`

	// Компании
	TotalCompanies   int `json:"total_companies"`
	PublicCompanies  int `json:"public_companies"`
	PrivateCompanies int `json:"private_companies"`

	// Связи
	TotalCompanyAccounts  int     `json:"total_company_accounts"`
	AvgAccountsPerCompany float64 `json:"avg_accounts_per_company"`
	MaxAccountsInCompany  int     `json:"max_accounts_in_company"`

	// Транзакции
	TotalTransactions   int `json:"total_transactions"`
	SuccessTransactions int `json:"success_transactions"`
	PendingTransactions int `json:"pending_transactions"`
	TrialTransactions   int `json:"trial_transactions"`

	// Активность
	ActiveCompanies7d  int `json:"active_companies_7d"`
	ActiveCompanies30d int `json:"active_companies_30d"`

	// Схемы
	CompanySchemasWithoutData int `json:"company_schemas_without_data"`
}
