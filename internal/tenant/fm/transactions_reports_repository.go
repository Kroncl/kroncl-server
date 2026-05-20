package fm

import (
	"context"
	"fmt"
	"time"
)

func (r *Repository) CreateTransactionReport(ctx context.Context, startDate, endDate time.Time, comment *string) (*TransactionsReport, int, error) {
	objectPath, total, err := r.GenerateTransactionsReport(ctx, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}

	query := `
		INSERT INTO transactions_reports (object_path, comment, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, object_path, comment, created_at, updated_at
	`

	var report TransactionsReport
	err = r.pool.QueryRow(ctx, query, objectPath, comment).Scan(
		&report.ID,
		&report.ObjectPath,
		&report.Comment,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to save report: %w", err)
	}

	return &report, total, nil
}

func (r *Repository) GetTransactionReports(ctx context.Context, offset, limit int, search *string) ([]TransactionsReport, int64, error) {
	var args []interface{}
	argIndex := 1

	countQuery := `SELECT COUNT(*) FROM transactions_reports`
	query := `
		SELECT id, object_path, comment, created_at, updated_at
		FROM transactions_reports
	`

	if search != nil && *search != "" {
		whereClause := fmt.Sprintf(" WHERE comment ILIKE $%d", argIndex)
		countQuery += whereClause
		query += whereClause
		args = append(args, "%"+*search+"%")
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get reports: %w", err)
	}
	defer rows.Close()

	var reports []TransactionsReport
	for rows.Next() {
		var report TransactionsReport
		err := rows.Scan(
			&report.ID,
			&report.ObjectPath,
			&report.Comment,
			&report.CreatedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, report)
	}

	return reports, total, nil
}
