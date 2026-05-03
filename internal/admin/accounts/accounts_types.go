package adminaccounts

type UserStats struct {
	TotalAccounts     int            `json:"total_accounts"`
	ConfirmedAccounts int            `json:"confirmed_accounts"`
	WaitingAccounts   int            `json:"waiting_accounts"`
	AdminAccounts     int            `json:"admin_accounts"`
	AccountsWithType  map[string]int `json:"accounts_with_type"`
}

type PromoteRequest struct {
	Level int `json:"level"`
}
