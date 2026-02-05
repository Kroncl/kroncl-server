package hrm

import "context"

func (r *Repository) RemoveEmployeeAccount(ctx context.Context, companyID, accountID string) error {
	// 1. Проверяем через companiesService
	err := r.companiesService.RemoveCompanyMember(ctx, companyID, accountID)
	if err != nil {
		return err
	}

	// 2. Удаляем связь если существует (без проверки результата)
	query := `
		DELETE FROM employee_account 
		WHERE account_id = $1
	`

	_, err = r.pool.Exec(ctx, query, accountID)

	return nil
}
