package storage

var ModuleTablesMap = map[string][]string{
	"hrm": {
		"employees",
		"employee_account",
		"employee_position",
		"employees_positions",
		"accounts_settings",
	},
	"fm": {
		"transactions",
		"transaction_categories",
		"transaction_category",
		"transaction_employee",
		"counterparties",
		"credits",
		"credit_transactions",
		"credit_counterparty",
		"deals_transactions",
	},
	"crm": {
		"clients",
		"client_source",
		"client_sources",
	},
	"wm": {
		"catalog_categories",
		"catalog_units",
		"catalog_unit_category",
		"stock_positions",
		"stock_batches",
		"stock_position_batch",
	},
	"dm": {
		"deals",
		"deal_types",
		"deal_statuses",
		"deal_status",
		"deal_employees",
		"deal_client",
		"deal_positions",
	},
}
