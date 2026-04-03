package support

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateTicket создаёт новый тикет с первым сообщением
func (s *Service) CreateTicket(ctx context.Context, companyID, accountID string, req *CreateTicketRequest) (*Ticket, error) {
	// Проверяем, что компания существует
	_, err := s.companiesService.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	// Проверяем количество активных тикетов
	var pendingCount int
	countQuery := `SELECT COUNT(*) FROM support_tickets WHERE company_id = $1 AND status = $2`
	err = s.pool.QueryRow(ctx, countQuery, companyID, TicketStatusPending).Scan(&pendingCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending tickets: %w", err)
	}
	if pendingCount >= config.SUPPORT_MAX_PENDING_TICKETS {
		return nil, fmt.Errorf("too many pending tickets. Maximum %d", config.SUPPORT_MAX_PENDING_TICKETS)
	}

	// Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Создаём тикет
	ticketID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO support_tickets (id, company_id, initiator_id, theme, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, company_id, initiator_id, theme, status, created_at, updated_at
	`

	var ticket Ticket
	err = tx.QueryRow(ctx, query,
		ticketID,
		companyID,
		accountID,
		req.Theme,
		TicketStatusPending,
		now,
		now,
	).Scan(
		&ticket.ID,
		&ticket.CompanyID,
		&ticket.InitiatorID,
		&ticket.Theme,
		&ticket.Status,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Создаём первое сообщение
	messageID := uuid.New().String()
	messageQuery := `
		INSERT INTO support_ticket_messages (id, account_id, ticket_id, text, read, is_tech, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = tx.Exec(ctx, messageQuery,
		messageID,
		accountID,
		ticketID,
		req.Text,
		true,
		false,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create initial message: %w", err)
	}

	// Коммитим
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Загружаем инициатора
	initiator, err := s.accountsService.GetPublicByID(ctx, ticket.InitiatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiator: %w", err)
	}
	ticket.Initiator = *initiator

	// Загружаем последнее сообщение
	lastMessage, err := s.getLastMessage(ctx, ticket.ID)
	if err == nil {
		ticket.LastMessage = lastMessage
	}

	return &ticket, nil
}

// GetTickets возвращает список тикетов компании с пагинацией
func (s *Service) GetTickets(ctx context.Context, companyID string, page, limit int) ([]Ticket, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Получаем общее количество
	var total int
	countQuery := `SELECT COUNT(*) FROM support_tickets WHERE company_id = $1`
	err := s.pool.QueryRow(ctx, countQuery, companyID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}

	// Получаем тикеты
	query := `
		SELECT id, company_id, initiator_id, theme, status, created_at, updated_at
		FROM support_tickets
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	var tickets []Ticket
	var initiatorIDs []string

	for rows.Next() {
		var t Ticket
		err := rows.Scan(
			&t.ID,
			&t.CompanyID,
			&t.InitiatorID,
			&t.Theme,
			&t.Status,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, t)
		initiatorIDs = append(initiatorIDs, t.InitiatorID)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	// Загружаем всех инициаторов одним запросом
	initiatorsMap, err := s.accountsService.GetPublicAccountsByIDs(ctx, initiatorIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get initiators: %w", err)
	}

	// Заполняем инициаторов
	for i := range tickets {
		if initiator, ok := initiatorsMap[tickets[i].InitiatorID]; ok {
			tickets[i].Initiator = initiator
		}
	}

	// Загружаем последние сообщения
	for i := range tickets {
		lastMsg, err := s.getLastMessage(ctx, tickets[i].ID)
		if err == nil {
			tickets[i].LastMessage = lastMsg
		}
	}

	return tickets, total, nil
}

// GetTicketByID возвращает один тикет с полными данными
func (s *Service) GetTicketByID(ctx context.Context, companyID, ticketID string) (*Ticket, error) {
	query := `
		SELECT id, company_id, initiator_id, theme, status, created_at, updated_at
		FROM support_tickets
		WHERE id = $1 AND company_id = $2
	`

	var ticket Ticket
	err := s.pool.QueryRow(ctx, query, ticketID, companyID).Scan(
		&ticket.ID,
		&ticket.CompanyID,
		&ticket.InitiatorID,
		&ticket.Theme,
		&ticket.Status,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Загружаем инициатора
	initiator, err := s.accountsService.GetPublicByID(ctx, ticket.InitiatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiator: %w", err)
	}
	ticket.Initiator = *initiator

	// Загружаем последнее сообщение одним запросом вместо всех сообщений
	lastMessage, err := s.getLastMessage(ctx, ticket.ID)
	if err == nil {
		ticket.LastMessage = lastMessage
	}

	return &ticket, nil
}

// UpdateTicketStatus обновляет статус тикета
// Можно обновить только с pending на closed или revoked
// revoked нельзя обновить (конечный статус)
func (s *Service) UpdateTicketStatus(ctx context.Context, companyID, ticketID string, status TicketStatus) (*Ticket, error) {
	// Получаем текущий статус тикета
	var currentStatus TicketStatus
	checkQuery := `
		SELECT status
		FROM support_tickets
		WHERE id = $1 AND company_id = $2
	`

	err := s.pool.QueryRow(ctx, checkQuery, ticketID, companyID).Scan(&currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Проверяем, что тикет не в конечном статусе
	if currentStatus == TicketStatusRevoked {
		return nil, fmt.Errorf("cannot update revoked ticket")
	}

	// Если пытаются установить revoked, проверяем что был pending
	if status == TicketStatusRevoked && currentStatus != TicketStatusPending {
		return nil, fmt.Errorf("only pending tickets can be revoked")
	}

	// Если пытаются установить closed, проверяем что был pending
	if status == TicketStatusClosed && currentStatus != TicketStatusPending {
		return nil, fmt.Errorf("only pending tickets can be closed")
	}

	// Обновляем статус
	updateQuery := `
		UPDATE support_tickets
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND company_id = $3
	`

	result, err := s.pool.Exec(ctx, updateQuery, status, ticketID, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("ticket not found")
	}

	// Возвращаем обновлённый тикет
	return s.GetTicketByID(ctx, companyID, ticketID)
}

// CheckTicketAccess проверяет, существует ли тикет и принадлежит ли он компании
func (s *Service) CheckTicketAccess(ctx context.Context, companyID, ticketID string) error {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM support_tickets
			WHERE id = $1 AND company_id = $2
		)
	`

	err := s.pool.QueryRow(ctx, query, ticketID, companyID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check ticket access: %w", err)
	}

	if !exists {
		return fmt.Errorf("ticket not found or access denied")
	}

	return nil
}
