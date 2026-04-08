package config

type CriticalityLevel int

const (
	CriticalityMin      CriticalityLevel = 1
	CriticalityLow      CriticalityLevel = 3
	CriticalityMedium   CriticalityLevel = 5
	CriticalityHigh     CriticalityLevel = 7
	CriticalityCritical CriticalityLevel = 9
	CriticalityMax      CriticalityLevel = 10
)

var PermissionCriticality = map[string]CriticalityLevel{
	// ========== COMPANY (3-10) ==========
	PERMISSION_COMPANY_UPDATE:         8, // изменение данных компании
	PERMISSION_PRICING_MIGRATE:        10,
	PERMISSION_PRICING_TRANSACTIONS:   3,
	PERMISSION_SUPPORT_TICKETS:        3,
	PERMISSION_SUPPORT_TICKETS_CREATE: 7,
	PERMISSION_SUPPORT_TICKETS_UPDATE: 8,
	PERMISSION_LOGS:                   3,
	PERMISSION_LOGS_CLEAR:             10,
	PERMISSION_LOGS_OPTIMIZE:          10,
	PERMISSION_LOGS_ACTIVITY:          7,

	// ========== STORAGE (1-2) ==========
	PERMISSION_STORAGE_SOURCES: 1, // просмотр ресурсов хранилища

	// ========== ACCOUNTS (1-10) ==========
	PERMISSION_ACCOUNTS:                    2,  // базовый доступ к аккаунтам
	PERMISSION_ACCOUNTS_DELETE:             10, // удаление аккаунта из компании
	PERMISSION_ACCOUNTS_INVITATIONS:        3,  // просмотр приглашений
	PERMISSION_ACCOUNTS_INVITATIONS_CREATE: 8,  // создание приглашения
	PERMISSION_ACCOUNTS_INVITATIONS_REVOKE: 8,  // отзыв приглашения

	// ========== HRM (2-9) ==========
	PERMISSION_HRM:                  2, // базовый доступ к HRM
	PERMISSION_HRM_EMPLOYEES:        3, // просмотр сотрудников
	PERMISSION_HRM_EMPLOYEES_CREATE: 8, // создание сотрудника
	PERMISSION_HRM_EMPLOYEES_UPDATE: 7, // обновление сотрудника
	PERMISSION_HRM_POSITIONS:        3,
	PERMISSION_HRM_POSITIONS_CREATE: 8,
	PERMISSION_HRM_POSITIONS_UPDATE: 7,
	PERMISSION_HRM_POSITIONS_DELETE: 9,
	PERMISSION_HRM_ANALYSIS:         7,

	// ========== FM (2-9) ==========
	PERMISSION_FM: 2, // базовый доступ к FM

	// ----- FM TRANSACTIONS (3-9) -----
	PERMISSION_FM_TRANSACTIONS:         3, // просмотр транзакций
	PERMISSION_FM_TRANSACTIONS_CREATE:  7, // создание транзакции
	PERMISSION_FM_TRANSACTIONS_REVERSE: 9, // сторно транзакции

	// ----- FM CATEGORIES (2-5) -----
	PERMISSION_FM_TRANSACTIONS_CATEGORIES:        2, // просмотр категорий
	PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE: 5, // создание категории
	PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE: 4, // обновление категории
	PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE: 5, // удаление категории

	// ----- FM ANALYSIS (1) -----
	PERMISSION_FM_ANALYSIS: 1, // просмотр аналитики

	// ----- FM COUNTERPARTIES (2-6) -----
	PERMISSION_FM_COUNTERPARTIES:        2, // просмотр контрагентов
	PERMISSION_FM_COUNTERPARTIES_CREATE: 6, // создание контрагента
	PERMISSION_FM_COUNTERPARTIES_UPDATE: 5, // обновление контрагента

	// ----- FM CREDITS (3-8) -----
	PERMISSION_FM_CREDITS:              3, // просмотр кредитов
	PERMISSION_FM_CREDITS_CREATE:       8, // создание кредита
	PERMISSION_FM_CREDITS_UPDATE:       7, // обновление кредита
	PERMISSION_FM_CREDITS_TRANSACTIONS: 4, // просмотр платежей по кредиту
	PERMISSION_FM_CREDITS_PAY:          8, // платёж по кредиту

	// ========== CRM (2-8) ==========
	PERMISSION_CRM:                2,
	PERMISSION_CRM_CLIENTS:        3,
	PERMISSION_CRM_CLIENTS_CREATE: 8,
	PERMISSION_CRM_CLIENTS_UPDATE: 7,
	PERMISSION_CRM_SOURCES:        3,
	PERMISSION_CRM_SOURCES_CREATE: 8,
	PERMISSION_CRM_SOURCES_UPDATE: 7,
	PERMISSION_CRM_ANALYSIS:       1,

	// ========== WM (2-8) ==========
	PERMISSION_WM:                           2,
	PERMISSION_WM_CATALOG:                   2,
	PERMISSION_WM_CATALOG_CATEGORIES:        2,
	PERMISSION_WM_CATALOG_CATEGORIES_CREATE: 8,
	PERMISSION_WM_CATALOG_CATEGORIES_UPDATE: 7,
	PERMISSION_WM_CATALOG_UNITS:             2,
	PERMISSION_WM_CATALOG_UNITS_CREATE:      8,
	PERMISSION_WM_CATALOG_UNITS_UPDATE:      7,
	PERMISSION_WM_STOCKS:                    2,
	PERMISSION_WM_STOCKS_BATCHES:            2,
	PERMISSION_WM_STOCKS_BATCHES_CREATE:     9,
	PERMISSION_WM_STOCKS_POSITIONS:          2,

	// ========== DM (2-9) ==========
	PERMISSION_DM:                 2,
	PERMISSION_DM_TYPES:           2,
	PERMISSION_DM_TYPES_CREATE:    4,
	PERMISSION_DM_TYPES_UPDATE:    5,
	PERMISSION_DM_TYPES_DELETE:    6,
	PERMISSION_DM_STATUSES:        2,
	PERMISSION_DM_STATUSES_CREATE: 4,
	PERMISSION_DM_STATUSES_UPDATE: 5,
	PERMISSION_DM_STATUSES_DELETE: 6,
	PERMISSION_DM_DEALS:           2,
	PERMISSION_DM_DEALS_CREATE:    8,
	PERMISSION_DM_DEALS_UPDATE:    7,
	PERMISSION_DM_DEALS_DELETE:    9,
	PERMISSION_DM_ANALYSIS:        1,
}

// GetCriticality returns criticality level for a permission
// Returns CriticalityMin (1) if permission not found
func GetCriticality(permission string) CriticalityLevel {
	if level, ok := PermissionCriticality[permission]; ok {
		return level
	}
	return CriticalityMin
}

// IsCritical checks if permission is critical (level >= 7)
func IsCritical(permission string) bool {
	return GetCriticality(permission) >= CriticalityHigh
}

// IsViewOnly checks if permission is view-only (level <= 3)
func IsViewOnly(permission string) bool {
	return GetCriticality(permission) <= CriticalityLow
}

// GetCriticalityLevels returns all permissions with their criticality
func GetCriticalityLevels() map[string]int {
	result := make(map[string]int, len(PermissionCriticality))
	for k, v := range PermissionCriticality {
		result[k] = int(v)
	}
	return result
}
