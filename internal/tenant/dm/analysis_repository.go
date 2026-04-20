package dm

import (
	"context"
	"fmt"
	"kroncl-server/internal/tenant/fm"
	"strconv"
	"strings"
)

// ---------
// ANALYSIS REPOSITORY
// ---------

func (r *Repository) GetDealAnalysisSummary(ctx context.Context, params GetAnalysisParams) (*DealAnalysisSummary, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	fromClause := `FROM deals d`
	whereClause := ""

	if params.StartDate != nil {
		conditions = append(conditions, "d.created_at >= $"+strconv.Itoa(argIndex))
		args = append(args, *params.StartDate)
		argIndex++
	}
	if params.EndDate != nil {
		conditions = append(conditions, "d.created_at <= $"+strconv.Itoa(argIndex))
		args = append(args, *params.EndDate)
		argIndex++
	}
	if params.TypeID != nil && *params.TypeID != "" {
		conditions = append(conditions, "d.type_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.TypeID)
		argIndex++
	}
	if params.StatusID != nil && *params.StatusID != "" {
		fromClause += ` INNER JOIN deal_status ds ON d.id = ds.deal_id`
		conditions = append(conditions, "ds.status_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.StatusID)
		argIndex++
	}
	if params.ClientID != nil && *params.ClientID != "" {
		fromClause += ` INNER JOIN deal_client dc ON d.id = dc.deal_id`
		conditions = append(conditions, "dc.client_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.ClientID)
		argIndex++
	}
	if params.EmployeeID != nil && *params.EmployeeID != "" {
		fromClause += ` INNER JOIN deal_employees de ON d.id = de.deal_id`
		conditions = append(conditions, "de.employee_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.EmployeeID)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var defaultStatusID, defaultStatusName *string
	err := r.pool.QueryRow(ctx, `SELECT id, name FROM deal_statuses WHERE is_default = true LIMIT 1`).Scan(&defaultStatusID, &defaultStatusName)
	if err != nil {
		defaultStatusID = nil
		defaultStatusName = nil
	}

	query := `
		SELECT 
			COUNT(d.id) as total_deals,
			COUNT(ds_default.deal_id) as deals_in_default
		` + fromClause + `
		LEFT JOIN deal_status ds_default ON d.id = ds_default.deal_id AND ds_default.status_id = $` + strconv.Itoa(argIndex) + `
	` + whereClause

	queryArgs := args
	if defaultStatusID != nil {
		queryArgs = append(queryArgs, *defaultStatusID)
	} else {
		queryArgs = append(queryArgs, nil)
	}

	var summary DealAnalysisSummary
	err = r.pool.QueryRow(ctx, query, queryArgs...).Scan(
		&summary.TotalDeals,
		&summary.DealsInDefault,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal analysis summary: %w", err)
	}

	summary.DefaultStatusID = defaultStatusID
	summary.DefaultStatusName = defaultStatusName

	if params.StartDate == nil && params.EndDate == nil && params.TypeID == nil && params.StatusID == nil && params.ClientID == nil && params.EmployeeID == nil {
		fmSummary, err := r.fmRepository.GetDealTransactionsSummary(ctx, "", fm.GetTransactionsRequest{})
		if err == nil && fmSummary.TotalCount > 0 {
			summary.AvgDealAmount = float64(fmSummary.TotalAmount) / float64(summary.TotalDeals)
		}
	}

	return &summary, nil
}

func (r *Repository) GetDealsGroupedByType(ctx context.Context, params GetAnalysisParams) ([]GroupedStats, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	fromClause := `FROM deals d LEFT JOIN deal_types dt ON d.type_id = dt.id`
	whereClause := ""

	if params.StartDate != nil {
		conditions = append(conditions, "d.created_at >= $"+strconv.Itoa(argIndex))
		args = append(args, *params.StartDate)
		argIndex++
	}
	if params.EndDate != nil {
		conditions = append(conditions, "d.created_at <= $"+strconv.Itoa(argIndex))
		args = append(args, *params.EndDate)
		argIndex++
	}
	if params.StatusID != nil && *params.StatusID != "" {
		fromClause += ` INNER JOIN deal_status ds ON d.id = ds.deal_id`
		conditions = append(conditions, "ds.status_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.StatusID)
		argIndex++
	}
	if params.ClientID != nil && *params.ClientID != "" {
		fromClause += ` INNER JOIN deal_client dc ON d.id = dc.deal_id`
		conditions = append(conditions, "dc.client_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.ClientID)
		argIndex++
	}
	if params.EmployeeID != nil && *params.EmployeeID != "" {
		fromClause += ` INNER JOIN deal_employees de ON d.id = de.deal_id`
		conditions = append(conditions, "de.employee_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.EmployeeID)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `
		SELECT 
			COALESCE(dt.id::text, 'no-type') as group_key,
			COALESCE(dt.name, 'Без типа') as group_name,
			COUNT(d.id) as count
	` + fromClause + whereClause + `
		GROUP BY dt.id, dt.name
		ORDER BY count DESC
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get deals grouped by type: %w", err)
	}
	defer rows.Close()

	var stats []GroupedStats
	for rows.Next() {
		var stat GroupedStats
		if err := rows.Scan(&stat.GroupKey, &stat.GroupName, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan grouped stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *Repository) GetDealsGroupedByStatus(ctx context.Context, params GetAnalysisParams) ([]GroupedStats, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	fromClause := `FROM deals d INNER JOIN deal_status ds ON d.id = ds.deal_id INNER JOIN deal_statuses s ON ds.status_id = s.id`
	whereClause := ""

	if params.StartDate != nil {
		conditions = append(conditions, "d.created_at >= $"+strconv.Itoa(argIndex))
		args = append(args, *params.StartDate)
		argIndex++
	}
	if params.EndDate != nil {
		conditions = append(conditions, "d.created_at <= $"+strconv.Itoa(argIndex))
		args = append(args, *params.EndDate)
		argIndex++
	}
	if params.TypeID != nil && *params.TypeID != "" {
		conditions = append(conditions, "d.type_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.TypeID)
		argIndex++
	}
	if params.ClientID != nil && *params.ClientID != "" {
		fromClause += ` INNER JOIN deal_client dc ON d.id = dc.deal_id`
		conditions = append(conditions, "dc.client_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.ClientID)
		argIndex++
	}
	if params.EmployeeID != nil && *params.EmployeeID != "" {
		fromClause += ` INNER JOIN deal_employees de ON d.id = de.deal_id`
		conditions = append(conditions, "de.employee_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.EmployeeID)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `
		SELECT 
			s.id as group_key,
			s.name as group_name,
			COUNT(DISTINCT d.id) as count
	` + fromClause + whereClause + `
		GROUP BY s.id, s.name, s.sort_order
		ORDER BY s.sort_order ASC
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get deals grouped by status: %w", err)
	}
	defer rows.Close()

	var stats []GroupedStats
	for rows.Next() {
		var stat GroupedStats
		if err := rows.Scan(&stat.GroupKey, &stat.GroupName, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan grouped stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *Repository) GetDealsGroupedByEmployee(ctx context.Context, params GetAnalysisParams) ([]GroupedStats, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	fromClause := `FROM deals d INNER JOIN deal_employees de ON d.id = de.deal_id INNER JOIN employees e ON de.employee_id = e.id`
	whereClause := ""

	if params.StartDate != nil {
		conditions = append(conditions, "d.created_at >= $"+strconv.Itoa(argIndex))
		args = append(args, *params.StartDate)
		argIndex++
	}
	if params.EndDate != nil {
		conditions = append(conditions, "d.created_at <= $"+strconv.Itoa(argIndex))
		args = append(args, *params.EndDate)
		argIndex++
	}
	if params.TypeID != nil && *params.TypeID != "" {
		conditions = append(conditions, "d.type_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.TypeID)
		argIndex++
	}
	if params.StatusID != nil && *params.StatusID != "" {
		fromClause += ` INNER JOIN deal_status ds ON d.id = ds.deal_id`
		conditions = append(conditions, "ds.status_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.StatusID)
		argIndex++
	}
	if params.ClientID != nil && *params.ClientID != "" {
		fromClause += ` INNER JOIN deal_client dc ON d.id = dc.deal_id`
		conditions = append(conditions, "dc.client_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.ClientID)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `
		SELECT 
			e.id as group_key,
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) as group_name,
			COUNT(DISTINCT d.id) as count
	` + fromClause + whereClause + `
		GROUP BY e.id, e.first_name, e.last_name
		ORDER BY count DESC
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get deals grouped by employee: %w", err)
	}
	defer rows.Close()

	var stats []GroupedStats
	for rows.Next() {
		var stat GroupedStats
		if err := rows.Scan(&stat.GroupKey, &stat.GroupName, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan grouped stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *Repository) GetDealsGroupedByClient(ctx context.Context, params GetAnalysisParams) ([]GroupedStats, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	fromClause := `FROM deals d INNER JOIN deal_client dc ON d.id = dc.deal_id INNER JOIN clients c ON dc.client_id = c.id`
	whereClause := ""

	if params.StartDate != nil {
		conditions = append(conditions, "d.created_at >= $"+strconv.Itoa(argIndex))
		args = append(args, *params.StartDate)
		argIndex++
	}
	if params.EndDate != nil {
		conditions = append(conditions, "d.created_at <= $"+strconv.Itoa(argIndex))
		args = append(args, *params.EndDate)
		argIndex++
	}
	if params.TypeID != nil && *params.TypeID != "" {
		conditions = append(conditions, "d.type_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.TypeID)
		argIndex++
	}
	if params.StatusID != nil && *params.StatusID != "" {
		fromClause += ` INNER JOIN deal_status ds ON d.id = ds.deal_id`
		conditions = append(conditions, "ds.status_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.StatusID)
		argIndex++
	}
	if params.EmployeeID != nil && *params.EmployeeID != "" {
		fromClause += ` INNER JOIN deal_employees de ON d.id = de.deal_id`
		conditions = append(conditions, "de.employee_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.EmployeeID)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `
		SELECT 
			c.id as group_key,
			CONCAT(c.first_name, ' ', COALESCE(c.last_name, '')) as group_name,
			COUNT(DISTINCT d.id) as count
	` + fromClause + whereClause + `
		GROUP BY c.id, c.first_name, c.last_name
		ORDER BY count DESC
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get deals grouped by client: %w", err)
	}
	defer rows.Close()

	var stats []GroupedStats
	for rows.Next() {
		var stat GroupedStats
		if err := rows.Scan(&stat.GroupKey, &stat.GroupName, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan grouped stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *Repository) GetDealsGroupedByTime(ctx context.Context, groupBy GroupBy, params GetAnalysisParams) ([]GroupedStats, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	fromClause := `FROM deals d`
	whereClause := ""
	var groupKeyFormat string

	switch groupBy {
	case GroupByDay:
		// timeFormat = "YYYY-MM-DD"
		groupKeyFormat = "TO_CHAR(DATE(d.created_at), 'YYYY-MM-DD')"
	case GroupByMonth:
		// timeFormat = "YYYY-MM"
		groupKeyFormat = "TO_CHAR(DATE_TRUNC('month', d.created_at), 'YYYY-MM')"
	case GroupByYear:
		// timeFormat = "YYYY"
		groupKeyFormat = "TO_CHAR(DATE_TRUNC('year', d.created_at), 'YYYY')"
	default:
		return nil, fmt.Errorf("invalid group_by for time: %s", groupBy)
	}

	if params.StartDate != nil {
		conditions = append(conditions, "d.created_at >= $"+strconv.Itoa(argIndex))
		args = append(args, *params.StartDate)
		argIndex++
	}
	if params.EndDate != nil {
		conditions = append(conditions, "d.created_at <= $"+strconv.Itoa(argIndex))
		args = append(args, *params.EndDate)
		argIndex++
	}
	if params.TypeID != nil && *params.TypeID != "" {
		conditions = append(conditions, "d.type_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.TypeID)
		argIndex++
	}
	if params.StatusID != nil && *params.StatusID != "" {
		fromClause += ` INNER JOIN deal_status ds ON d.id = ds.deal_id`
		conditions = append(conditions, "ds.status_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.StatusID)
		argIndex++
	}
	if params.ClientID != nil && *params.ClientID != "" {
		fromClause += ` INNER JOIN deal_client dc ON d.id = dc.deal_id`
		conditions = append(conditions, "dc.client_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.ClientID)
		argIndex++
	}
	if params.EmployeeID != nil && *params.EmployeeID != "" {
		fromClause += ` INNER JOIN deal_employees de ON d.id = de.deal_id`
		conditions = append(conditions, "de.employee_id = $"+strconv.Itoa(argIndex))
		args = append(args, *params.EmployeeID)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as group_key,
			%s as group_name,
			COUNT(d.id) as count
		%s %s
		GROUP BY group_key, group_name
		ORDER BY group_key ASC
	`, groupKeyFormat, groupKeyFormat, fromClause, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get deals grouped by time: %w", err)
	}
	defer rows.Close()

	var stats []GroupedStats
	for rows.Next() {
		var stat GroupedStats
		if err := rows.Scan(&stat.GroupKey, &stat.GroupName, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan grouped stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *Repository) GetDealsFinancialSummary(ctx context.Context, params GetAnalysisParams) (*fm.DealTransactionsSummary, error) {
	return r.fmRepository.GetOverallDealTransactionsSummary(ctx, fm.GetTransactionsRequest{
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
	})
}
