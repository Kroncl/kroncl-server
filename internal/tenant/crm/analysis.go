package crm

import (
	"context"
	"fmt"
	"time"
)

type GroupBy string

const (
	GroupBySource GroupBy = "source"
	GroupByDay    GroupBy = "day"
	GroupByMonth  GroupBy = "month"
)

type GroupedStats struct {
	GroupKey     string `json:"group_key"`     // source_id, date
	GroupName    string `json:"group_name"`    // source_name, date
	ClientsCount int64  `json:"clients_count"` // количество клиентов
}

// ClientsSummary represents summary statistics for a period
type ClientsSummary struct {
	TotalClients      int64 `json:"total_clients"`
	ActiveClients     int64 `json:"active_clients"`
	InactiveClients   int64 `json:"inactive_clients"`
	IndividualClients int64 `json:"individual_clients"`
	LegalClients      int64 `json:"legal_clients"`
	NewClients        int64 `json:"new_clients"` // клиенты за период
}

// GetGroupedClients returns client statistics grouped by source/day/month
func (r *Repository) GetGroupedClients(ctx context.Context,
	groupBy GroupBy,
	startDate, endDate *time.Time) ([]GroupedStats, error) {

	var selectCols, groupByExpr, joinClause string

	switch groupBy {
	case GroupBySource:
		selectCols = `
			cs.id::text as group_key,
			cs.name as group_name`
		groupByExpr = `cs.id, cs.name`
		joinClause = `RIGHT JOIN client_source csl ON c.id = csl.client_id
		              RIGHT JOIN client_sources cs ON csl.source_id = cs.id`

	case GroupByDay:
		selectCols = `
			DATE(c.created_at)::text as group_key,
			TO_CHAR(DATE(c.created_at), 'DD.MM.YYYY') as group_name`
		groupByExpr = `DATE(c.created_at)`
		joinClause = ``

	case GroupByMonth:
		selectCols = `
			TO_CHAR(DATE_TRUNC('month', c.created_at), 'YYYY-MM') as group_key,
			TO_CHAR(DATE_TRUNC('month', c.created_at), 'MMMM YYYY') as group_name`
		groupByExpr = `DATE_TRUNC('month', c.created_at)`
		joinClause = ``

	default:
		return nil, fmt.Errorf("invalid group_by: %s", groupBy)
	}

	query := fmt.Sprintf(`
		SELECT 
			%s,
			COUNT(DISTINCT c.id) as clients_count
		FROM clients c
		%s
		WHERE ($1::timestamptz IS NULL OR c.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR c.created_at <= $2)
		GROUP BY %s
		ORDER BY clients_count DESC, group_name ASC
	`, selectCols, joinClause, groupByExpr)

	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query grouped clients: %w", err)
	}
	defer rows.Close()

	var stats []GroupedStats
	for rows.Next() {
		var stat GroupedStats
		err := rows.Scan(
			&stat.GroupKey,
			&stat.GroupName,
			&stat.ClientsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grouped stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetClientsSummary returns summary statistics for given period
func (r *Repository) GetClientsSummary(ctx context.Context, startDate, endDate *time.Time) (*ClientsSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_clients,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_clients,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_clients,
			COUNT(CASE WHEN type = 'individual' THEN 1 END) as individual_clients,
			COUNT(CASE WHEN type = 'legal' THEN 1 END) as legal_clients,
			COUNT(CASE 
				WHEN ($1::timestamptz IS NULL OR created_at >= $1)
				 AND ($2::timestamptz IS NULL OR created_at <= $2)
				THEN 1 
			END) as new_clients
		FROM clients
	`

	var summary ClientsSummary
	err := r.pool.QueryRow(ctx, query, startDate, endDate).Scan(
		&summary.TotalClients,
		&summary.ActiveClients,
		&summary.InactiveClients,
		&summary.IndividualClients,
		&summary.LegalClients,
		&summary.NewClients,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get clients summary: %w", err)
	}

	return &summary, nil
}
