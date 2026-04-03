package support

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreateMessage создаёт новое сообщение в тикете
func (s *Service) CreateMessage(ctx context.Context, ticketID, accountID, text string) (*Message, error) {
	checkQuery := `
		SELECT is_tech
		FROM support_ticket_messages
		WHERE ticket_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := s.pool.Query(ctx, checkQuery, ticketID, config.SUPPORT_MAX_MESSAGES_IN_ROW)
	if err != nil {
		return nil, fmt.Errorf("failed to check last messages: %w", err)
	}
	defer rows.Close()

	var nonTechCount int
	for rows.Next() {
		var isTech bool
		if err := rows.Scan(&isTech); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		if !isTech {
			nonTechCount++
		} else {
			break
		}
	}

	if nonTechCount >= config.SUPPORT_MAX_MESSAGES_IN_ROW {
		return nil, fmt.Errorf("you have %d consecutive messages without support response. Please wait for support to reply", config.SUPPORT_MAX_MESSAGES_IN_ROW)
	}

	messageID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO support_ticket_messages (id, account_id, ticket_id, text, read, is_tech, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, account_id, ticket_id, text, read, is_tech, created_at, updated_at
	`

	var msg Message
	err = s.pool.QueryRow(ctx, query,
		messageID,
		accountID,
		ticketID,
		text,
		true,
		false,
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

	// Загружаем аккаунт
	account, err := s.accountsService.GetPublicByID(ctx, msg.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	msg.Account = *account

	return &msg, nil
}

// getMessagesWithAccounts возвращает все сообщения тикета с аккаунтами и ссылками
func (s *Service) GetMessagesWithAccounts(ctx context.Context, ticketID string) ([]Message, error) {
	// Получаем сообщения
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

	var messages []Message
	var accountIDs []string

	for rows.Next() {
		var msg Message
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Загружаем аккаунты одним запросом
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

	// Загружаем ссылки для всех сообщений
	linksMap, err := s.getLinksForMessages(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}

	// Заполняем ссылки
	for i := range messages {
		if links, ok := linksMap[messages[i].ID]; ok {
			messages[i].Links = links
		}
	}

	return messages, nil
}

// getLastMessage возвращает последнее сообщение тикета
func (s *Service) getLastMessage(ctx context.Context, ticketID string) (*Message, error) {
	query := `
		SELECT id, account_id, ticket_id, text, read, is_tech, created_at, updated_at
		FROM support_ticket_messages
		WHERE ticket_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var msg Message
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

// getLinksForMessages возвращает ссылки для списка сообщений
func (s *Service) getLinksForMessages(ctx context.Context, messages []Message) (map[string][]Link, error) {
	if len(messages) == 0 {
		return make(map[string][]Link), nil
	}

	// Собираем ID сообщений
	messageIDs := make([]string, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.ID
	}

	// Создаём плейсхолдеры
	placeholders := make([]string, len(messageIDs))
	args := make([]interface{}, len(messageIDs))
	for i, id := range messageIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, message_id, link, capture, created_at, updated_at
		FROM support_ticket_message_links
		WHERE message_id IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}
	defer rows.Close()

	linksMap := make(map[string][]Link)
	for rows.Next() {
		var link Link
		err := rows.Scan(
			&link.ID,
			&link.MessageID,
			&link.Link,
			&link.Capture,
			&link.CreatedAt,
			&link.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		linksMap[link.MessageID] = append(linksMap[link.MessageID], link)
	}

	return linksMap, nil
}

// GetMessages возвращает все сообщения тикета с пагинацией (от новых к старым)
func (s *Service) GetMessages(ctx context.Context, ticketID string, page, limit int) ([]Message, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Получаем общее количество сообщений
	var total int
	countQuery := `SELECT COUNT(*) FROM support_ticket_messages WHERE ticket_id = $1`
	err := s.pool.QueryRow(ctx, countQuery, ticketID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Получаем сообщения с пагинацией (от новых к старым)
	query := `
		SELECT id, account_id, ticket_id, text, read, is_tech, created_at, updated_at
		FROM support_ticket_messages
		WHERE ticket_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, ticketID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	var accountIDs []string

	for rows.Next() {
		var msg Message
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
			return nil, 0, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
		accountIDs = append(accountIDs, msg.AccountID)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	// Загружаем аккаунты одним запросом
	accountsMap, err := s.accountsService.GetPublicAccountsByIDs(ctx, accountIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get accounts: %w", err)
	}

	// Заполняем аккаунты
	for i := range messages {
		if account, ok := accountsMap[messages[i].AccountID]; ok {
			messages[i].Account = account
		}
	}

	// Загружаем ссылки для всех сообщений
	linksMap, err := s.getLinksForMessages(ctx, messages)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get links: %w", err)
	}

	// Заполняем ссылки
	for i := range messages {
		if links, ok := linksMap[messages[i].ID]; ok {
			messages[i].Links = links
		}
	}

	return messages, total, nil
}

// UpdateMessageReadStatus обновляет статус прочтения сообщения и возвращает обновлённое сообщение
func (s *Service) UpdateMessageReadStatus(ctx context.Context, messageID string, read bool) (*Message, error) {
	updateQuery := `
		UPDATE support_ticket_messages
		SET read = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, account_id, ticket_id, text, read, is_tech, created_at, updated_at
	`

	var msg Message
	err := s.pool.QueryRow(ctx, updateQuery, read, messageID).Scan(
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
		return nil, fmt.Errorf("failed to update message read status: %w", err)
	}

	// Загружаем аккаунт
	account, err := s.accountsService.GetPublicByID(ctx, msg.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	msg.Account = *account

	return &msg, nil
}

// CheckMessageAccess проверяет, принадлежит ли сообщение тикету
func (s *Service) CheckMessageAccess(ctx context.Context, ticketID, messageID string) error {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM support_ticket_messages
			WHERE id = $1 AND ticket_id = $2
		)
	`

	err := s.pool.QueryRow(ctx, query, messageID, ticketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check message access: %w", err)
	}

	if !exists {
		return fmt.Errorf("message not found or does not belong to ticket")
	}

	return nil
}
