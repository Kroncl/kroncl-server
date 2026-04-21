package config

const (
	PERMISSION_SUPPORT_TICKETS                   = "support.tickets"
	PERMISSION_SUPPORT_TICKETS_CREATE            = "support.tickets.create"
	PERMISSION_SUPPORT_TICKETS_UPDATE            = "support.tickets.update"
	PERMISSION_PRICING_MIGRATE                   = "pricing.migrate"
	PERMISSION_PRICING_TRANSACTIONS              = "pricing.transactions"
	PERMISSION_COMPANY_UPDATE                    = "company.update"
	PERMISSION_COMPANY_DELETE                    = "company.delete"
	PERMISSION_STORAGE_SOURCES                   = "storage.sources"
	PERMISSION_LOGS                              = "logs"
	PERMISSION_LOGS_CLEAR                        = "logs.clear"
	PERMISSION_LOGS_OPTIMIZE                     = "logs.optimize"
	PERMISSION_LOGS_ACTIVITY                     = "logs.activity"
	PERMISSION_ACCOUNTS                          = "accounts"
	PERMISSION_ACCOUNTS_DELETE                   = "accounts.delete"
	PERMISSION_ACCOUNTS_SETTINGS                 = "accounts.settings"
	PERMISSION_ACCOUNTS_SETTINGS_UPDATE          = "accounts.settings.update"
	PERMISSION_ACCOUNTS_INVITATIONS              = "accounts.invitations"
	PERMISSION_ACCOUNTS_INVITATIONS_CREATE       = "accounts.invitations.create"
	PERMISSION_ACCOUNTS_INVITATIONS_REVOKE       = "accounts.invitations.revoke"
	PERMISSION_HRM                               = "hrm"
	PERMISSION_HRM_EMPLOYEES                     = "hrm.employees"
	PERMISSION_HRM_EMPLOYEES_CREATE              = "hrm.employees.create"
	PERMISSION_HRM_EMPLOYEES_UPDATE              = "hrm.employees.update"
	PERMISSION_HRM_POSITIONS                     = "hrm.positions"
	PERMISSION_HRM_POSITIONS_CREATE              = "hrm.positions.create"
	PERMISSION_HRM_POSITIONS_UPDATE              = "hrm.positions.update"
	PERMISSION_HRM_POSITIONS_DELETE              = "hrm.positions.delete"
	PERMISSION_HRM_ANALYSIS                      = "hrm.analysis"
	PERMISSION_FM                                = "fm"
	PERMISSION_FM_TRANSACTIONS                   = "fm.transactions"
	PERMISSION_FM_TRANSACTIONS_CREATE            = "fm.transactions.create"
	PERMISSION_FM_TRANSACTIONS_REVERSE           = "fm.transactions.reverse"
	PERMISSION_FM_TRANSACTIONS_CATEGORIES        = "fm.transactions.categories"
	PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE = "fm.transactions.categories.create"
	PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE = "fm.transactions.categories.update"
	PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE = "fm.transactions.categories.delete"
	PERMISSION_FM_ANALYSIS                       = "fm.analysis"
	PERMISSION_FM_COUNTERPARTIES                 = "fm.counterparties"
	PERMISSION_FM_COUNTERPARTIES_CREATE          = "fm.counterparties.create"
	PERMISSION_FM_COUNTERPARTIES_UPDATE          = "fm.counterparties.update"
	PERMISSION_FM_CREDITS                        = "fm.credits"
	PERMISSION_FM_CREDITS_CREATE                 = "fm.credits.create"
	PERMISSION_FM_CREDITS_UPDATE                 = "fm.credits.update"
	PERMISSION_FM_CREDITS_TRANSACTIONS           = "fm.credits.transactions"
	PERMISSION_FM_CREDITS_PAY                    = "fm.credits.pay"
	PERMISSION_CRM                               = "crm"
	PERMISSION_CRM_CLIENTS                       = "crm.clients"
	PERMISSION_CRM_CLIENTS_CREATE                = "crm.clients.create"
	PERMISSION_CRM_CLIENTS_UPDATE                = "crm.clients.update"
	PERMISSION_CRM_SOURCES                       = "crm.sources"
	PERMISSION_CRM_SOURCES_CREATE                = "crm.sources.create"
	PERMISSION_CRM_SOURCES_UPDATE                = "crm.sources.update"
	PERMISSION_CRM_ANALYSIS                      = "crm.analysis"
	PERMISSION_WM                                = "wm"
	PERMISSION_WM_CATALOG                        = "wm.catalog"
	PERMISSION_WM_CATALOG_CATEGORIES             = "wm.catalog.categories"
	PERMISSION_WM_CATALOG_CATEGORIES_CREATE      = "wm.catalog.categories.create"
	PERMISSION_WM_CATALOG_CATEGORIES_UPDATE      = "wm.catalog.categories.update"
	PERMISSION_WM_CATALOG_UNITS                  = "wm.catalog.units"
	PERMISSION_WM_CATALOG_UNITS_CREATE           = "wm.catalog.units.create"
	PERMISSION_WM_CATALOG_UNITS_UPDATE           = "wm.catalog.units.update"
	PERMISSION_WM_STOCKS                         = "wm.stocks"
	PERMISSION_WM_STOCKS_BATCHES                 = "wm.stocks.batches"
	PERMISSION_WM_STOCKS_BATCHES_CREATE          = "wm.stocks.batches.create"
	PERMISSION_WM_STOCKS_POSITIONS               = "wm.stocks.positions"
	PERMISSION_DM                                = "dm"
	PERMISSION_DM_TYPES                          = "dm.types"
	PERMISSION_DM_TYPES_CREATE                   = "dm.types.create"
	PERMISSION_DM_TYPES_UPDATE                   = "dm.types.update"
	PERMISSION_DM_TYPES_DELETE                   = "dm.types.delete"
	PERMISSION_DM_STATUSES                       = "dm.statuses"
	PERMISSION_DM_STATUSES_CREATE                = "dm.statuses.create"
	PERMISSION_DM_STATUSES_UPDATE                = "dm.statuses.update"
	PERMISSION_DM_STATUSES_DELETE                = "dm.statuses.delete"
	PERMISSION_DM_DEALS                          = "dm.deals"
	PERMISSION_DM_DEALS_CREATE                   = "dm.deals.create"
	PERMISSION_DM_DEALS_UPDATE                   = "dm.deals.update"
	PERMISSION_DM_DEALS_DELETE                   = "dm.deals.delete"
	PERMISSION_DM_DEALS_TRANSACTIONS             = "dm.deals.transactions"
	PERMISSION_DM_DEALS_TRANSACTIONS_CREATE      = "dm.deals.transactions.create"
	PERMISSION_DM_DEALS_TRANSACTIONS_SUMMARY     = "dm.deals.transactions.summary"
	PERMISSION_DM_ANALYSIS                       = "dm.analysis"
)

// --------
// UTILS
// --------

// GetAllPermissions возвращает список всех существующих разрешений
func GetAllPermissions() []string {
	return []string{
		PERMISSION_SUPPORT_TICKETS,
		PERMISSION_SUPPORT_TICKETS_CREATE,
		PERMISSION_SUPPORT_TICKETS_UPDATE,
		PERMISSION_PRICING_MIGRATE,
		PERMISSION_PRICING_TRANSACTIONS,
		PERMISSION_COMPANY_UPDATE,
		PERMISSION_COMPANY_DELETE,
		PERMISSION_STORAGE_SOURCES,
		PERMISSION_LOGS,
		PERMISSION_LOGS_CLEAR,
		PERMISSION_LOGS_OPTIMIZE,
		PERMISSION_LOGS_ACTIVITY,
		PERMISSION_ACCOUNTS,
		PERMISSION_ACCOUNTS_DELETE,
		PERMISSION_ACCOUNTS_INVITATIONS,
		PERMISSION_ACCOUNTS_INVITATIONS_CREATE,
		PERMISSION_ACCOUNTS_INVITATIONS_REVOKE,
		PERMISSION_ACCOUNTS_SETTINGS,
		PERMISSION_ACCOUNTS_SETTINGS_UPDATE,
		PERMISSION_HRM,
		PERMISSION_HRM_EMPLOYEES,
		PERMISSION_HRM_EMPLOYEES_CREATE,
		PERMISSION_HRM_EMPLOYEES_UPDATE,
		PERMISSION_HRM_POSITIONS,
		PERMISSION_HRM_POSITIONS_CREATE,
		PERMISSION_HRM_POSITIONS_UPDATE,
		PERMISSION_HRM_POSITIONS_DELETE,
		PERMISSION_HRM_ANALYSIS,
		PERMISSION_FM,
		PERMISSION_FM_TRANSACTIONS,
		PERMISSION_FM_TRANSACTIONS_CREATE,
		PERMISSION_FM_TRANSACTIONS_REVERSE,
		PERMISSION_FM_TRANSACTIONS_CATEGORIES,
		PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE,
		PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE,
		PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE,
		PERMISSION_FM_ANALYSIS,
		PERMISSION_FM_COUNTERPARTIES,
		PERMISSION_FM_COUNTERPARTIES_CREATE,
		PERMISSION_FM_COUNTERPARTIES_UPDATE,
		PERMISSION_FM_CREDITS,
		PERMISSION_FM_CREDITS_CREATE,
		PERMISSION_FM_CREDITS_UPDATE,
		PERMISSION_FM_CREDITS_TRANSACTIONS,
		PERMISSION_FM_CREDITS_PAY,
		PERMISSION_CRM,
		PERMISSION_CRM_CLIENTS,
		PERMISSION_CRM_CLIENTS_CREATE,
		PERMISSION_CRM_CLIENTS_UPDATE,
		PERMISSION_CRM_SOURCES,
		PERMISSION_CRM_SOURCES_CREATE,
		PERMISSION_CRM_SOURCES_UPDATE,
		PERMISSION_CRM_ANALYSIS,
		PERMISSION_WM,
		PERMISSION_WM_CATALOG,
		PERMISSION_WM_CATALOG_CATEGORIES,
		PERMISSION_WM_CATALOG_CATEGORIES_CREATE,
		PERMISSION_WM_CATALOG_CATEGORIES_UPDATE,
		PERMISSION_WM_CATALOG_UNITS,
		PERMISSION_WM_CATALOG_UNITS_CREATE,
		PERMISSION_WM_CATALOG_UNITS_UPDATE,
		PERMISSION_WM_STOCKS,
		PERMISSION_WM_STOCKS_BATCHES,
		PERMISSION_WM_STOCKS_BATCHES_CREATE,
		PERMISSION_WM_STOCKS_POSITIONS,
		PERMISSION_DM,
		PERMISSION_DM_TYPES,
		PERMISSION_DM_TYPES_CREATE,
		PERMISSION_DM_TYPES_UPDATE,
		PERMISSION_DM_TYPES_DELETE,
		PERMISSION_DM_STATUSES,
		PERMISSION_DM_STATUSES_CREATE,
		PERMISSION_DM_STATUSES_UPDATE,
		PERMISSION_DM_STATUSES_DELETE,
		PERMISSION_DM_DEALS,
		PERMISSION_DM_DEALS_CREATE,
		PERMISSION_DM_DEALS_UPDATE,
		PERMISSION_DM_DEALS_DELETE,
		PERMISSION_DM_DEALS_TRANSACTIONS,
		PERMISSION_DM_DEALS_TRANSACTIONS_CREATE,
		PERMISSION_DM_DEALS_TRANSACTIONS_SUMMARY,
		PERMISSION_DM_ANALYSIS,
	}
}

// GetGuestPermissions возвращает разрешения, доступные гостевой роли (RoleGuest)
// Гость имеет только права на чтение базовых данных, без возможности создания/изменения
func GetGuestPermissions() map[string]bool {
	return map[string]bool{
		PERMISSION_SUPPORT_TICKETS:        true,
		PERMISSION_SUPPORT_TICKETS_CREATE: true,
		PERMISSION_SUPPORT_TICKETS_UPDATE: true,
		PERMISSION_PRICING_TRANSACTIONS:   true,
		PERMISSION_STORAGE_SOURCES:        true,
		PERMISSION_LOGS:                   true,
		PERMISSION_LOGS_ACTIVITY:          true,
		PERMISSION_ACCOUNTS:               true,
		PERMISSION_ACCOUNTS_INVITATIONS:   true,

		// modules
		PERMISSION_HRM:               true,
		PERMISSION_HRM_EMPLOYEES:     true,
		PERMISSION_HRM_POSITIONS:     true,
		PERMISSION_FM:                true,
		PERMISSION_FM_TRANSACTIONS:   true,
		PERMISSION_FM_COUNTERPARTIES: true,
		PERMISSION_FM_CREDITS:        true,

		PERMISSION_CRM:         true,
		PERMISSION_CRM_CLIENTS: true,
		PERMISSION_CRM_SOURCES: true,

		PERMISSION_WM:                    true,
		PERMISSION_WM_CATALOG:            true,
		PERMISSION_WM_CATALOG_CATEGORIES: true,
		PERMISSION_WM_CATALOG_UNITS:      true,
		PERMISSION_WM_STOCKS:             true,

		PERMISSION_DM:                    true,
		PERMISSION_DM_TYPES:              true,
		PERMISSION_DM_STATUSES:           true,
		PERMISSION_DM_DEALS:              true,
		PERMISSION_DM_DEALS_TRANSACTIONS: true,
	}
}

// IsGuestAllowed проверяет, доступно ли разрешение для гостевой роли
func IsGuestAllowed(permission string) bool {
	allowed := GetGuestPermissions()
	return allowed[permission]
}

// IsValidPermission проверяет, существует ли разрешение в списке доступных
func IsValidPermission(permission string) bool {
	for _, p := range GetAllPermissions() {
		if p == permission {
			return true
		}
	}
	return false
}

// ValidatePermissions проверяет список разрешений и возвращает невалидные
func ValidatePermissions(permissions []string) (invalid []string) {
	for _, p := range permissions {
		if !IsValidPermission(p) {
			invalid = append(invalid, p)
		}
	}
	return
}

// UniquePermissions удаляет дубликаты из списка разрешений
func UniquePermissions(permissions []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, p := range permissions {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result
}
