package hrm

import (
	"context"
	"fmt"
	"time"
)

// GetEmployeesSummary возвращает суммарную аналитику по сотрудникам за период
func (r *Repository) GetEmployeesSummary(ctx context.Context, startDate, endDate *time.Time) (*EmployeesSummary, error) {
	// Базовые условия для фильтрации по дате создания
	var dateFilter string
	var args []interface{}
	argIndex := 1

	if startDate != nil {
		dateFilter += fmt.Sprintf(" AND e.created_at >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}
	if endDate != nil {
		dateFilter += fmt.Sprintf(" AND e.created_at <= $%d", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	// Запрос для суммарной статистики
	query := fmt.Sprintf(`
		SELECT 
			(SELECT COUNT(*) FROM employees_positions) as total_positions,
			(SELECT COUNT(*) FROM employees) as total_employees,
			(SELECT COUNT(*) FROM employees WHERE status = 'active') as active_employees,
			(SELECT COUNT(*) FROM employees WHERE status = 'inactive') as inactive_employees,
			COALESCE((SELECT COUNT(*) FROM employees WHERE 1=1 %s), 0) as new_employees
	`, dateFilter)

	var summary EmployeesSummary
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&summary.TotalPositions,
		&summary.TotalEmployees,
		&summary.ActiveEmployees,
		&summary.InactiveEmployees,
		&summary.NewEmployees,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get employees summary: %w", err)
	}

	return &summary, nil
}

// GetEmployeesGrouped возвращает динамику изменения штата сотрудников по дням/месяцам/годам
func (r *Repository) GetEmployeesGrouped(ctx context.Context, startDate, endDate *time.Time, groupBy GroupBy) ([]GroupedStats, error) {
	// Определяем формат группировки
	var dateFormat string
	switch groupBy {
	case GroupByDay:
		dateFormat = "DATE(e.created_at)"
	case GroupByMonth:
		dateFormat = "DATE_TRUNC('month', e.created_at)"
	case GroupByYear:
		dateFormat = "DATE_TRUNC('year', e.created_at)"
	default:
		dateFormat = "DATE(e.created_at)"
	}

	// Базовые условия для фильтрации
	var dateFilter string
	var args []interface{}
	argIndex := 1

	if startDate != nil {
		dateFilter += fmt.Sprintf(" AND e.created_at >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}
	if endDate != nil {
		dateFilter += fmt.Sprintf(" AND e.created_at <= $%d", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	// Запрос для группировки
	query := fmt.Sprintf(`
		SELECT 
			TO_CHAR(%s, 'YYYY-MM-DD') as group_key,
			TO_CHAR(%s, 'YYYY-MM-DD') as group_name,
			COUNT(*) as employees_count,
			SUM(CASE WHEN e.status = 'active' THEN 1 ELSE 0 END) as active_count,
			SUM(CASE WHEN e.status = 'inactive' THEN 1 ELSE 0 END) as inactive_count
		FROM employees e
		WHERE 1=1 %s
		GROUP BY %s
		ORDER BY %s ASC
	`, dateFormat, dateFormat, dateFilter, dateFormat, dateFormat)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get employees grouped: %w", err)
	}
	defer rows.Close()

	var stats []GroupedStats
	for rows.Next() {
		var s GroupedStats
		err := rows.Scan(
			&s.GroupKey,
			&s.GroupName,
			&s.EmployeesCount,
			&s.ActiveCount,
			&s.InactiveCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grouped stats: %w", err)
		}
		stats = append(stats, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stats, nil
}
