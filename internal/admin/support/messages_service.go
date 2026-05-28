package adminsupport

import (
	"context"
	"fmt"
	"kroncl-server/internal/mailer"
	"kroncl-server/internal/tenant/support"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetTicketMessages возвращает все сообщения тикета (от старых к новым)
func (s *Service) GetTicketMessages(ctx context.Context, ticketID string) ([]support.Message, error) {
	// Проверяем существование тикета
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM support_tickets WHERE id = $1)`
	err := s.pool.QueryRow(ctx, checkQuery, ticketID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check ticket: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("ticket not found")
	}

	// Получаем все сообщения
	query := `
        SELECT id, account_id, ticket_id, text, read, is_tech, created_at, updated_at
        FROM support_ticket_messages
        WHERE ticket_id = $1
        ORDER BY created_at ASC
    `

	rows, err := s.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []support.Message
	var accountIDs []string

	for rows.Next() {
		var msg support.Message
		err := rows.Scan(
			&msg.ID,
			&msg.AccountID,
			&msg.TicketID,
			&msg.Text,
			&msg.Read,
			&msg.IsTech,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
		accountIDs = append(accountIDs, msg.AccountID)
	}

	// Загружаем аккаунты
	accountsMap, err := s.accountsService.GetPublicAccountsByIDs(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	// Заполняем аккаунты
	for i := range messages {
		if account, ok := accountsMap[messages[i].AccountID]; ok {
			messages[i].Account = account
		}
	}

	return messages, nil
}

// CreateAdminMessage создаёт новое сообщение от админа (требует назначения на тикет)
func (s *Service) CreateAdminMessage(ctx context.Context, ticketID, adminID, text string) (*support.Message, error) {
	// Проверяем, что админ назначен на этот тикет
	var assignedAdminID string
	checkQuery := `SELECT admin_id FROM support_tickets_admins WHERE ticket_id = $1`
	err := s.pool.QueryRow(ctx, checkQuery, ticketID).Scan(&assignedAdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not assigned to any admin")
		}
		return nil, fmt.Errorf("failed to check assigned admin: %w", err)
	}

	if assignedAdminID != adminID {
		return nil, fmt.Errorf("you are not assigned to this ticket")
	}

	// Проверяем, что тикет не закрыт
	var status support.TicketStatus
	statusQuery := `SELECT status FROM support_tickets WHERE id = $1`
	err = s.pool.QueryRow(ctx, statusQuery, ticketID).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket status: %w", err)
	}
	if status != support.TicketStatusPending {
		return nil, fmt.Errorf("cannot add message to closed or revoked ticket")
	}

	// Получаем информацию о тикете и компании
	var companyID string
	var initiatorID string
	ticketQuery := `SELECT company_id, initiator_id FROM support_tickets WHERE id = $1`
	err = s.pool.QueryRow(ctx, ticketQuery, ticketID).Scan(&companyID, &initiatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket info: %w", err)
	}

	// Получаем компанию
	_, err = s.companiesService.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	// Получаем аккаунт инициатора (клиента)
	initiatorAccount, err := s.accountsService.GetPublicByID(ctx, initiatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiator account: %w", err)
	}

	// Создаём сообщение
	messageID := uuid.New().String()
	now := time.Now()

	insertQuery := `
        INSERT INTO support_ticket_messages (id, account_id, ticket_id, text, read, is_tech, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, account_id, ticket_id, text, read, is_tech, created_at, updated_at
    `

	var msg support.Message
	err = s.pool.QueryRow(ctx, insertQuery,
		messageID,
		adminID,
		ticketID,
		text,
		false, // не прочитано (для клиента)
		true,  // is_tech = true (сообщение от техподдержки)
		now,
		now,
	).Scan(
		&msg.ID,
		&msg.AccountID,
		&msg.TicketID,
		&msg.Text,
		&msg.Read,
		&msg.IsTech,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Загружаем аккаунт админа
	account, err := s.accountsService.GetPublicByID(ctx, msg.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	msg.Account = *account

	// Обновляем updated_at тикета
	updateTicketQuery := `UPDATE support_tickets SET updated_at = NOW() WHERE id = $1`
	_, err = s.pool.Exec(ctx, updateTicketQuery, ticketID)
	if err != nil {
		// не критично, логируем
		fmt.Printf("failed to update ticket updated_at: %v\n", err)
	}

	// Асинхронно отправляем уведомление клиенту
	go func() {
		s.mailer.SendSupportAnswer(context.Background(), &mailer.ClientSupportMessageData{
			ClientEmail: initiatorAccount.Email,
			ClientName:  initiatorAccount.Name,
			Message:     text,
			TicketID:    ticketID,
		})
	}()

	return &msg, nil
}

// UpdateAdminMessage обновляет текст сообщения админа (только свои, и только если тикет не закрыт)
func (s *Service) UpdateAdminMessage(ctx context.Context, messageID, adminID, newText string) (*support.Message, error) {
	// Получаем сообщение и проверяем права
	var msg support.Message
	var ticketID string
	var isTech bool
	var messageAccountID string

	getQuery := `
        SELECT id, account_id, ticket_id, text, is_tech, created_at, updated_at
        FROM support_ticket_messages
        WHERE id = $1
    `
	err := s.pool.QueryRow(ctx, getQuery, messageID).Scan(
		&msg.ID,
		&messageAccountID,
		&ticketID,
		&msg.Text,
		&isTech,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Проверяем, что сообщение от админа (is_tech = true)
	if !isTech {
		return nil, fmt.Errorf("cannot edit non-admin message")
	}

	// Проверяем, что текущий админ — автор сообщения
	if messageAccountID != adminID {
		return nil, fmt.Errorf("you can only edit your own messages")
	}

	// Проверяем, что тикет не закрыт
	var status support.TicketStatus
	statusQuery := `SELECT status FROM support_tickets WHERE id = $1`
	err = s.pool.QueryRow(ctx, statusQuery, ticketID).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket status: %w", err)
	}
	if status != support.TicketStatusPending {
		return nil, fmt.Errorf("cannot edit message in closed or revoked ticket")
	}

	// Обновляем сообщение
	updateQuery := `
        UPDATE support_ticket_messages
        SET text = $1, updated_at = NOW()
        WHERE id = $2
        RETURNING id, account_id, ticket_id, text, read, is_tech, created_at, updated_at
    `
	err = s.pool.QueryRow(ctx, updateQuery, newText, messageID).Scan(
		&msg.ID,
		&msg.AccountID,
		&msg.TicketID,
		&msg.Text,
		&msg.Read,
		&msg.IsTech,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Загружаем аккаунт
	account, err := s.accountsService.GetPublicByID(ctx, msg.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	msg.Account = *account

	return &msg, nil
}

// DeleteAdminMessage удаляет сообщение админа (только свои)
func (s *Service) DeleteAdminMessage(ctx context.Context, messageID, adminID string) error {
	// Получаем сообщение и проверяем права
	var messageAccountID string
	var ticketID string
	var isTech bool

	getQuery := `
        SELECT account_id, ticket_id, is_tech
        FROM support_ticket_messages
        WHERE id = $1
    `
	err := s.pool.QueryRow(ctx, getQuery, messageID).Scan(&messageAccountID, &ticketID, &isTech)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Проверяем, что сообщение от админа
	if !isTech {
		return fmt.Errorf("cannot delete non-admin message")
	}

	// Проверяем, что текущий админ — автор сообщения
	if messageAccountID != adminID {
		return fmt.Errorf("you can only delete your own messages")
	}

	// Проверяем, что тикет не закрыт
	var status support.TicketStatus
	statusQuery := `SELECT status FROM support_tickets WHERE id = $1`
	err = s.pool.QueryRow(ctx, statusQuery, ticketID).Scan(&status)
	if err != nil {
		return fmt.Errorf("failed to get ticket status: %w", err)
	}
	if status != support.TicketStatusPending {
		return fmt.Errorf("cannot delete message in closed or revoked ticket")
	}

	// Удаляем сообщение
	deleteQuery := `DELETE FROM support_ticket_messages WHERE id = $1`
	_, err = s.pool.Exec(ctx, deleteQuery, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (s *Service) getLastMessageByTicketID(ctx context.Context, ticketID string) (*support.Message, error) {
	query := `
		SELECT id, account_id, ticket_id, text, read, is_tech, created_at, updated_at
		FROM support_ticket_messages
		WHERE ticket_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var msg support.Message
	err := s.pool.QueryRow(ctx, query, ticketID).Scan(
		&msg.ID,
		&msg.AccountID,
		&msg.TicketID,
		&msg.Text,
		&msg.Read,
		&msg.IsTech,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Загружаем аккаунт
	account, err := s.accountsService.GetPublicByID(ctx, msg.AccountID)
	if err != nil {
		return nil, err
	}
	msg.Account = *account

	return &msg, nil
}
