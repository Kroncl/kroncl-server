package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// GetCompanyInvitations возвращает приглашения в компанию с пагинацией и фильтрацией
func (s *Service) GetCompanyInvitations(
	ctx context.Context,
	companyID string,
	params GetInvitationsRequest,
) (*GetInvitationsResponse, error) {
	// Базовый запрос
	baseQuery := `
        SELECT 
            id, email, company_id, status,
            created_at, updated_at
        FROM company_invitations
        WHERE company_id = $1
    `

	// Запрос для подсчета общего количества
	countQuery := `
        SELECT COUNT(*) 
        FROM company_invitations
        WHERE company_id = $1
    `

	// Подготавливаем аргументы
	args := []interface{}{companyID}
	argCounter := 2 // $1 уже занят company_id

	// Добавляем фильтр по статусу
	if params.Status != "" {
		// Валидируем статус
		validStatuses := map[string]bool{
			InvitationStatusWaiting:  true,
			InvitationStatusAccepted: true,
			InvitationStatusRejected: true,
		}
		if !validStatuses[params.Status] {
			return nil, fmt.Errorf("invalid status filter. Allowed values: waiting, accepted, rejected")
		}

		whereCondition := ` AND status = $` + strconv.Itoa(argCounter)
		baseQuery += whereCondition
		countQuery += whereCondition
		args = append(args, params.Status)
		argCounter++
	}

	// Добавляем поиск по email если есть
	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		whereCondition := ` AND LOWER(email) LIKE $` + strconv.Itoa(argCounter)
		baseQuery += whereCondition
		countQuery += whereCondition
		args = append(args, searchPattern)
		argCounter++
	}

	// Получаем общее количество для пагинации
	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count invitations: %w", err)
	}

	// Добавляем сортировку и лимиты для основного запроса
	baseQuery += " ORDER BY created_at DESC"
	baseQuery += " LIMIT $" + strconv.Itoa(argCounter) + " OFFSET $" + strconv.Itoa(argCounter+1)
	args = append(args, params.Limit, params.Offset)

	// Выполняем основной запрос
	rows, err := s.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query invitations: %w", err)
	}
	defer rows.Close()

	// Собираем результаты
	var invitations []CompanyInvitation
	for rows.Next() {
		var invitation CompanyInvitation
		err := rows.Scan(
			&invitation.ID,
			&invitation.Email,
			&invitation.CompanyID,
			&invitation.Status,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Создаем пагинацию
	pagination := core.NewPagination(total, params.Page, params.Limit)

	return &GetInvitationsResponse{
		Invitations: invitations,
		Pagination:  pagination,
	}, nil
}

// HasPendingInvitation проверяет есть ли ожидающие приглашения для email и компании
func (s *Service) HasPendingInvitation(
	ctx context.Context,
	email string,
	companyID string,
) (bool, *CompanyInvitation, error) {
	query := `
        SELECT 
            id, email, company_id, status,
            created_at, updated_at
        FROM company_invitations
        WHERE email = $1 
            AND company_id = $2 
            AND status = $3
        LIMIT 1
    `

	var invitation CompanyInvitation
	err := s.pool.QueryRow(
		ctx,
		query,
		strings.ToLower(email),
		companyID,
		InvitationStatusWaiting,
	).Scan(
		&invitation.ID,
		&invitation.Email,
		&invitation.CompanyID,
		&invitation.Status,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check pending invitation: %w", err)
	}

	return true, &invitation, nil
}

// CreateInvitationAtomic создает приглашение с использованием транзакции для атомарности
func (s *Service) CreateInvitationAtomic(
	ctx context.Context,
	companyID string,
	invitedBy string,
	req *CreateInvitationRequest,
) (*InvitationResponse, error) {
	// Валидация email
	if !isValidEmail(req.Email) {
		return nil, fmt.Errorf("invalid email format")
	}

	normalizedEmail := strings.ToLower(req.Email)

	// Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	// 1. Проверяем существование компании
	var companyExists bool
	err = tx.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`,
		companyID,
	).Scan(&companyExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check company existence: %w", err)
	}
	if !companyExists {
		return nil, fmt.Errorf("company not found")
	}

	// 2. Проверяем, является ли пользователь с email уже членом компании
	var isAlreadyMember bool
	err = tx.QueryRow(
		ctx,
		`SELECT EXISTS(
            SELECT 1 
            FROM accounts a
            INNER JOIN company_accounts ca ON a.id = ca.account_id
            WHERE LOWER(a.email) = $1 
                AND a.status = 'confirmed'
                AND ca.company_id = $2
        )`,
		normalizedEmail,
		companyID,
	).Scan(&isAlreadyMember)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if isAlreadyMember {
		return nil, fmt.Errorf("user with email %s is already a member of this company", req.Email)
	}

	// 3. Проверяем существующие приглашения
	var hasPendingInvitation bool
	var existingInvitationID string
	err = tx.QueryRow(
		ctx,
		`SELECT id, true 
         FROM company_invitations 
         WHERE email = $1 
            AND company_id = $2 
            AND status = $3 
         LIMIT 1`,
		normalizedEmail,
		companyID,
		InvitationStatusWaiting,
	).Scan(&existingInvitationID, &hasPendingInvitation)

	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to check pending invitations: %w", err)
	}

	if hasPendingInvitation {
		var invitation CompanyInvitation
		err = tx.QueryRow(
			ctx,
			`SELECT id, email, company_id, status, created_at, updated_at
             FROM company_invitations 
             WHERE id = $1`,
			existingInvitationID,
		).Scan(
			&invitation.ID,
			&invitation.Email,
			&invitation.CompanyID,
			&invitation.Status,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing invitation: %w", err)
		}

		tx.Rollback(ctx)
		tx = nil

		return &InvitationResponse{
			Invitation: invitation,
			Message:    "Invitation already exists and is pending",
		}, nil
	}

	var invitation CompanyInvitation
	err = tx.QueryRow(
		ctx,
		`INSERT INTO company_invitations (
            email, company_id, status
        ) VALUES ($1, $2, $3)
        RETURNING id, email, company_id, status, created_at, updated_at`,
		normalizedEmail,
		companyID,
		InvitationStatusWaiting,
	).Scan(
		&invitation.ID,
		&invitation.Email,
		&invitation.CompanyID,
		&invitation.Status,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			// Конкурентное создание - откатываемся и проверяем снова
			tx.Rollback(ctx)
			tx = nil

			// Проверяем еще раз вне транзакции
			hasPending, existingInvitation, checkErr := s.HasPendingInvitation(ctx, normalizedEmail, companyID)
			if checkErr == nil && hasPending {
				return &InvitationResponse{
					Invitation: *existingInvitation,
					Message:    "Invitation was created by another process",
				}, nil
			}
			return nil, fmt.Errorf("concurrent invitation creation detected")
		}
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// 5. Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	tx = nil

	return &InvitationResponse{
		Invitation: invitation,
		Message:    "Invitation created successfully",
	}, nil
}

// UpdateInvitationStatus обновляет статус приглашения
func (s *Service) UpdateInvitationStatus(
	ctx context.Context,
	invitationID string,
	status string,
) (*CompanyInvitation, error) {
	// Валидация статуса
	validStatuses := map[string]bool{
		InvitationStatusWaiting:  true,
		InvitationStatusAccepted: true,
		InvitationStatusRejected: true,
	}
	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status. Allowed values: waiting, accepted, rejected")
	}

	// Для принятия приглашения нужны дополнительные проверки
	if status == InvitationStatusAccepted {
		return s.acceptInvitation(ctx, invitationID)
	}

	// Обновление статуса на waiting или rejected
	query := `
        UPDATE company_invitations 
        SET status = $1, updated_at = NOW()
        WHERE id = $2
        RETURNING id, email, company_id, status, created_at, updated_at
    `

	var invitation CompanyInvitation
	err := s.pool.QueryRow(
		ctx,
		query,
		status,
		invitationID,
	).Scan(
		&invitation.ID,
		&invitation.Email,
		&invitation.CompanyID,
		&invitation.Status,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, fmt.Errorf("failed to update invitation status: %w", err)
	}

	return &invitation, nil
}

// acceptInvitation обрабатывает принятие приглашения
func (s *Service) acceptInvitation(ctx context.Context, invitationID string) (*CompanyInvitation, error) {
	// Используем транзакцию для атомарности
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	// 1. Получаем информацию о приглашении
	var invitation CompanyInvitation
	var accountID string

	err = tx.QueryRow(
		ctx,
		`SELECT 
            ci.id, ci.email, ci.company_id, ci.status, 
            ci.created_at, ci.updated_at,
            a.id as account_id
        FROM company_invitations ci
        LEFT JOIN accounts a ON LOWER(a.email) = LOWER(ci.email) AND a.status = 'confirmed'
        WHERE ci.id = $1`,
		invitationID,
	).Scan(
		&invitation.ID,
		&invitation.Email,
		&invitation.CompanyID,
		&invitation.Status,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&accountID,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// 2. Проверяем, что приглашение в статусе waiting
	if invitation.Status != InvitationStatusWaiting {
		return nil, fmt.Errorf("invitation is not in waiting status")
	}

	// 3. Проверяем, существует ли аккаунт
	if accountID == "" {
		return nil, fmt.Errorf("account with email %s not found or not confirmed", invitation.Email)
	}

	// 4. Проверяем, не является ли пользователь уже членом компании
	var isAlreadyMember bool
	err = tx.QueryRow(
		ctx,
		`SELECT EXISTS(
            SELECT 1 
            FROM company_accounts 
            WHERE account_id = $1 
                AND company_id = $2
        )`,
		accountID,
		invitation.CompanyID,
	).Scan(&isAlreadyMember)

	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	if isAlreadyMember {
		return nil, fmt.Errorf("user is already a member of this company")
	}

	// 5. Обновляем статус приглашения
	_, err = tx.Exec(
		ctx,
		`UPDATE company_invitations 
         SET status = $1, updated_at = NOW()
         WHERE id = $2`,
		InvitationStatusAccepted,
		invitationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update invitation status: %w", err)
	}

	// 6. Добавляем пользователя в компанию (с ролью member по умолчанию)
	// Получаем ID роли member
	var roleID int
	err = tx.QueryRow(
		ctx,
		`SELECT id FROM roles WHERE code = $1`,
		RoleMember,
	).Scan(&roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find member role: %w", err)
	}

	// Добавляем пользователя в компанию
	_, err = tx.Exec(
		ctx,
		`INSERT INTO company_accounts (
            company_id, account_id, role_id, permissions,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, NOW(), NOW())
        ON CONFLICT (company_id, account_id) DO NOTHING`,
		invitation.CompanyID,
		accountID,
		roleID,
		`{}`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to company: %w", err)
	}

	// 7. Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	tx = nil

	// 8. Возвращаем обновленное приглашение
	invitation.Status = InvitationStatusAccepted
	// Обновляем updated_at (он обновился в БД, но не в нашей структуре)
	// Можно сделать дополнительный запрос или просто вернуть как есть
	return &invitation, nil
}

// WithdrawInvitation отзывает (полностью удаляет) приглашение
func (s *Service) WithdrawInvitation(
	ctx context.Context,
	invitationID string,
) error {
	// Простое удаление без дополнительных проверок
	query := `DELETE FROM company_invitations WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, invitationID)
	if err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}

	// Проверяем, была ли удалена запись
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found")
	}

	return nil
}

// WithdrawInvitationByEmail отзывает приглашение по email и company_id
func (s *Service) WithdrawInvitationByEmail(
	ctx context.Context,
	email string,
	companyID string,
) error {
	normalizedEmail := strings.ToLower(email)

	query := `
        DELETE FROM company_invitations 
        WHERE LOWER(email) = $1 AND company_id = $2 AND status = $3
    `

	result, err := s.pool.Exec(
		ctx,
		query,
		normalizedEmail,
		companyID,
		InvitationStatusWaiting,
	)
	if err != nil {
		return fmt.Errorf("failed to delete invitation by email: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no pending invitation found for email %s in company %s", email, companyID)
	}

	return nil
}

// GetInvitationByID получает приглашение по ID (вспомогательный метод)
func (s *Service) GetInvitationByID(
	ctx context.Context,
	invitationID string,
) (*CompanyInvitation, error) {
	query := `
        SELECT 
            id, email, company_id, status,
            created_at, updated_at
        FROM company_invitations
        WHERE id = $1
    `

	var invitation CompanyInvitation
	err := s.pool.QueryRow(ctx, query, invitationID).Scan(
		&invitation.ID,
		&invitation.Email,
		&invitation.CompanyID,
		&invitation.Status,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return &invitation, nil
}

// ValidateInvitationStatus валидирует статус приглашения (вспомогательная функция)
func ValidateInvitationStatus(status string) error {
	validStatuses := map[string]bool{
		InvitationStatusWaiting:  true,
		InvitationStatusAccepted: true,
		InvitationStatusRejected: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid invitation status: %s. Allowed values: waiting, accepted, rejected", status)
	}

	return nil
}

func isValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}

	at := strings.LastIndex(email, "@")
	if at < 1 || at > len(email)-4 {
		return false
	}

	dot := strings.LastIndex(email[at:], ".")
	if dot < 2 || dot > len(email[at:])-3 {
		return false
	}

	return true
}

// GetInvitationsByEmail возвращает все приглашения для указанного email
func (s *Service) GetInvitationsByEmail(
	ctx context.Context,
	email string,
	params GetInvitationsByEmailRequest,
) (*GetInvitationsByEmailResponse, error) {
	// Нормализуем email
	normalizedEmail := strings.ToLower(email)

	// Базовый запрос
	baseQuery := `
        SELECT 
            ci.id, ci.email, ci.company_id, ci.status,
            ci.created_at, ci.updated_at,
            c.name as company_name,
            c.avatar_url as company_avatar_url
        FROM company_invitations ci
        LEFT JOIN companies c ON ci.company_id = c.id
        WHERE ci.email = $1
    `

	// Запрос для подсчета общего количества
	countQuery := `
        SELECT COUNT(*) 
        FROM company_invitations
        WHERE email = $1
    `

	// Подготавливаем аргументы
	args := []interface{}{normalizedEmail}
	argCounter := 2 // $1 уже занят email

	// Добавляем фильтр по статусу если указан
	if params.Status != "" {
		// Валидируем статус
		validStatuses := map[string]bool{
			InvitationStatusWaiting:  true,
			InvitationStatusAccepted: true,
			InvitationStatusRejected: true,
		}
		if !validStatuses[params.Status] {
			return nil, fmt.Errorf("invalid status filter. Allowed values: waiting, accepted, rejected")
		}

		whereCondition := ` AND ci.status = $` + strconv.Itoa(argCounter)
		baseQuery += whereCondition
		countQuery += ` AND status = $` + strconv.Itoa(argCounter)
		args = append(args, params.Status)
		argCounter++
	}

	// Получаем общее количество для пагинации
	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count invitations: %w", err)
	}

	// Добавляем сортировку и лимиты для основного запроса
	baseQuery += " ORDER BY ci.created_at DESC"
	baseQuery += " LIMIT $" + strconv.Itoa(argCounter) + " OFFSET $" + strconv.Itoa(argCounter+1)
	args = append(args, params.Limit, params.Offset)

	// Выполняем основной запрос
	rows, err := s.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query invitations: %w", err)
	}
	defer rows.Close()

	// Собираем результаты
	var invitations []InvitationWithCompany
	for rows.Next() {
		var invitation InvitationWithCompany
		err := rows.Scan(
			&invitation.ID,
			&invitation.Email,
			&invitation.CompanyID,
			&invitation.Status,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
			&invitation.CompanyName,
			&invitation.CompanyAvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Создаем пагинацию
	pagination := core.NewPagination(total, params.Page, params.Limit)

	return &GetInvitationsByEmailResponse{
		Invitations: invitations,
		Pagination:  pagination,
	}, nil
}
