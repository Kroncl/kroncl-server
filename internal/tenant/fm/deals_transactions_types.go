package fm

const (
	DEAL_TRANSACTION_CATEGORY_INCOME_SLUG  = "deal-income"
	DEAL_TRANSACTION_CATEGORY_EXPENSE_SLUG = "deal-expense"
)

type DealTransactionsSummary struct {
	TotalAmount   int64 `json:"total_amount"`   // итоговая сумма (доходы - расходы)
	IncomeAmount  int64 `json:"income_amount"`  // сумма доходных транзакций
	ExpenseAmount int64 `json:"expense_amount"` // сумма расходных транзакций
	IncomeCount   int64 `json:"income_count"`   // количество доходных транзакций
	ExpenseCount  int64 `json:"expense_count"`  // количество расходных транзакций
	TotalCount    int64 `json:"total_count"`    // общее количество транзакций
}
