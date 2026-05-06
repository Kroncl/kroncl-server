package adminsupport

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/support"
	"strings"

	"github.com/jackc/pgx/v5"
)

// GetAllTickets возвращает все тикеты с фильтрацией и пагинацией
func (s *Service) GetAllTickets(ctx context.Context, status *support.TicketStatus, params core.PaginationParams) ([]AdminTicket, core.Pagination, error) {
	baseQuery := `
		SELECT 
			t.id, t.company_id, t.initiator_id, t.theme, t.status, t.created_at, t.updated_at,
			ta.admin_id as assigned_admin_id
		FROM support_tickets t
		LEFT JOIN support_tickets_admins ta ON t.id = ta.ticket_id
	`

	countQuery := `SELECT COUNT(*) FROM support_tickets t`

	var args []interface{}
	var whereClauses []string
	argCounter := 1

	if status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.status = $%d", argCounter))
		args = append(args, *status)
		argCounter++
	}

	if len(whereClauses) > 0 {
		where := " WHERE " + strings.Join(whereClauses, " AND ")
		baseQuery += where
		countQuery += where
	}

	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to count tickets: %w", err)
	}

	baseQuery += " ORDER BY t.created_at DESC"
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := s.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	var tickets []AdminTicket
	var companyIDs []string
	var initiatorIDs []string
	var ticketIDs []string

	for rows.Next() {
		var at AdminTicket

		var assignedAdminID *string
		err := rows.Scan(
			&at.ID,
			&at.Company.ID,
			&at.Initiator.ID,
			&at.Theme,
			&at.Status,
			&at.CreatedAt,
			&at.UpdatedAt,
			&assignedAdminID, // теперь это *string
		)
		if err != nil {
			return nil, core.Pagination{}, fmt.Errorf("failed to scan ticket: %w", err)
		}
		if assignedAdminID != nil {
			at.AssignedAdminID = assignedAdminID
		}
		tickets = append(tickets, at)
		companyIDs = append(companyIDs, at.Company.ID)
		initiatorIDs = append(initiatorIDs, at.Initiator.ID)
		ticketIDs = append(ticketIDs, at.ID)
	}

	// Загружаем компании
	companiesMap, err := s.companiesService.GetCompaniesByIDs(ctx, companyIDs)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to get companies: %w", err)
	}

	// Загружаем инициаторов
	initiatorsMap, err := s.accountsService.GetPublicAccountsByIDs(ctx, initiatorIDs)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to get initiators: %w", err)
	}

	// Загружаем последние сообщения (прямой запрос)
	lastMessagesMap := make(map[string]*support.Message)
	for _, ticketID := range ticketIDs {
		lastMsg, err := s.getLastMessageByTicketID(ctx, ticketID)
		if err == nil {
			lastMessagesMap[ticketID] = lastMsg
		}
	}

	// Заполняем данные
	for i := range tickets {
		if company, ok := companiesMap[tickets[i].Company.ID]; ok {
			tickets[i].Company = company
		}
		if initiator, ok := initiatorsMap[tickets[i].Initiator.ID]; ok {
			tickets[i].Initiator = initiator
		}
		if msg, ok := lastMessagesMap[tickets[i].ID]; ok {
			tickets[i].LastMessage = msg
		}
	}

	pagination := core.NewPagination(total, params.Page, params.Limit)

	return tickets, pagination, nil
}

// GetTicketByID возвращает один тикет с полной информацией
func (s *Service) GetTicketByID(ctx context.Context, ticketID string) (*AdminTicket, error) {
	query := `
		SELECT 
			t.id, t.company_id, t.initiator_id, t.theme, t.status, t.created_at, t.updated_at,
			ta.admin_id as assigned_admin_id
		FROM support_tickets t
		LEFT JOIN support_tickets_admins ta ON t.id = ta.ticket_id
		WHERE t.id = $1
	`

	var at AdminTicket
	var assignedAdminID *string
	err := s.pool.QueryRow(ctx, query, ticketID).Scan(
		&at.ID,
		&at.Company.ID,
		&at.Initiator.ID,
		&at.Theme,
		&at.Status,
		&at.CreatedAt,
		&at.UpdatedAt,
		&assignedAdminID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	at.AssignedAdminID = assignedAdminID
	// Загружаем компанию
	company, err := s.companiesService.GetCompanyByID(ctx, at.Company.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	at.Company = *company

	// Загружаем инициатора
	initiator, err := s.accountsService.GetPublicByID(ctx, at.Initiator.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiator: %w", err)
	}
	at.Initiator = *initiator

	// Загружаем последнее сообщение
	lastMessage, err := s.getLastMessageByTicketID(ctx, ticketID)
	if err == nil {
		at.LastMessage = lastMessage
	}

	return &at, nil
}

// UpdateTicketStatus обновляет статус тикета
func (s *Service) UpdateTicketStatus(ctx context.Context, ticketID string, status support.TicketStatus) error {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM support_tickets WHERE id = $1)`
	err := s.pool.QueryRow(ctx, checkQuery, ticketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check ticket: %w", err)
	}
	if !exists {
		return fmt.Errorf("ticket not found")
	}

	updateQuery := `
		UPDATE support_tickets
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err = s.pool.Exec(ctx, updateQuery, status, ticketID)
	if err != nil {
		return fmt.Errorf("failed to update ticket status: %w", err)
	}

	return nil
}

// adminsupport/service.go - добавить методы

// UnassignTicket отвязывает текущего админа от тикета
func (s *Service) UnassignTicket(ctx context.Context, ticketID, adminID string) error {
	// Проверяем, что тикет существует
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM support_tickets WHERE id = $1)`
	err := s.pool.QueryRow(ctx, checkQuery, ticketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check ticket: %w", err)
	}
	if !exists {
		return fmt.Errorf("ticket not found")
	}

	// Проверяем, что текущий админ действительно назначен на этот тикет
	var assignedAdminID string
	getAdminQuery := `SELECT admin_id FROM support_tickets_admins WHERE ticket_id = $1`
	err = s.pool.QueryRow(ctx, getAdminQuery, ticketID).Scan(&assignedAdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("no admin assigned to this ticket")
		}
		return fmt.Errorf("failed to check assigned admin: %w", err)
	}

	if assignedAdminID != adminID {
		return fmt.Errorf("you are not assigned to this ticket")
	}

	// Удаляем назначение
	deleteQuery := `DELETE FROM support_tickets_admins WHERE ticket_id = $1 AND admin_id = $2`
	_, err = s.pool.Exec(ctx, deleteQuery, ticketID, adminID)
	if err != nil {
		return fmt.Errorf("failed to unassign ticket: %w", err)
	}

	return nil
}

// AssignTicketWithCheck назначает админа на тикет с проверкой, что не назначен другой
func (s *Service) AssignTicketWithCheck(ctx context.Context, ticketID, adminID string) error {
	// Проверяем, есть ли уже назначенный админ
	var assignedAdminID *string
	checkQuery := `SELECT admin_id FROM support_tickets_admins WHERE ticket_id = $1`
	err := s.pool.QueryRow(ctx, checkQuery, ticketID).Scan(&assignedAdminID)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("failed to check assigned admin: %w", err)
	}

	if assignedAdminID != nil && *assignedAdminID != "" {
		return fmt.Errorf("ticket already assigned to another admin")
	}

	// Назначаем
	query := `
        INSERT INTO support_tickets_admins (ticket_id, admin_id)
        VALUES ($1, $2)
        ON CONFLICT (ticket_id) DO UPDATE SET
            admin_id = EXCLUDED.admin_id,
            updated_at = NOW()
    `
	_, err = s.pool.Exec(ctx, query, ticketID, adminID)
	if err != nil {
		return fmt.Errorf("failed to assign admin to ticket: %w", err)
	}
	return nil
}
