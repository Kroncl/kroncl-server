package accounts

import (
	"context"
	"fmt"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/core"
	"strings"
)

// GetAccountInvitationsWithPagination возвращает приглашения с пагинацией
func (s *Service) GetAccountInvitations(
	ctx context.Context,
	accountID string,
	params companies.GetInvitationsByEmailRequest,
) (*companies.GetInvitationsByEmailResponse, error) {
	// 1. Получаем аккаунт по ID
	account, err := s.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// 2. Проверяем статус аккаунта
	if account.Status != "confirmed" {
		// Для неподтвержденных аккаунтов возвращаем пустой список
		return &companies.GetInvitationsByEmailResponse{
			Invitations: []companies.InvitationWithCompany{},
			Pagination:  core.NewPagination(0, params.Page, params.Limit),
		}, nil
	}

	// 3. Получаем приглашения через сервис компаний
	response, err := s.companiesService.GetInvitationsByEmail(ctx, account.Email, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitations: %w", err)
	}

	return response, nil
}

// AcceptInvitation принимает приглашение от имени аккаунта
func (s *Service) AcceptInvitation(
	ctx context.Context,
	accountID string,
	invitationID string,
) (*companies.CompanyInvitation, error) {
	// 1. Получаем аккаунт по ID
	account, err := s.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// 2. Проверяем статус аккаунта
	if account.Status != "confirmed" {
		return nil, fmt.Errorf("account must be confirmed to accept invitations")
	}

	// 3. Получаем приглашение
	invitation, err := s.companiesService.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// 4. Проверяем, что email приглашения совпадает с email аккаунта
	if strings.ToLower(invitation.Email) != strings.ToLower(account.Email) {
		return nil, fmt.Errorf("invitation does not belong to this account")
	}

	// 5. Обновляем статус приглашения
	updatedInvitation, err := s.companiesService.UpdateInvitationStatus(ctx, invitationID, companies.InvitationStatusAccepted)
	if err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	return updatedInvitation, nil
}

// RejectInvitation отклоняет приглашение от имени аккаунта
func (s *Service) RejectInvitation(
	ctx context.Context,
	accountID string,
	invitationID string,
) (*companies.CompanyInvitation, error) {
	// 1. Получаем аккаунт по ID
	account, err := s.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// 2. Проверяем статус аккаунта
	if account.Status != "confirmed" {
		return nil, fmt.Errorf("account must be confirmed to reject invitations")
	}

	// 3. Получаем приглашение
	invitation, err := s.companiesService.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// 4. Проверяем, что email приглашения совпадает с email аккаунта
	if strings.ToLower(invitation.Email) != strings.ToLower(account.Email) {
		return nil, fmt.Errorf("invitation does not belong to this account")
	}

	// 5. Обновляем статус приглашения
	updatedInvitation, err := s.companiesService.UpdateInvitationStatus(ctx, invitationID, companies.InvitationStatusRejected)
	if err != nil {
		return nil, fmt.Errorf("failed to reject invitation: %w", err)
	}

	return updatedInvitation, nil
}
