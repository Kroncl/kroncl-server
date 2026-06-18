package billing

import (
	"bytes"
	"context"
	"fmt"

	"github.com/rentifly/tinkoff"
)

func (s *Service) InitPayment(ctx context.Context, req *InitPaymentRequest) (*InitPaymentResponse, error) {
	innerReq := &tinkoff.InitRequest{
		Amount:          req.Amount,
		OrderID:         req.OrderID,
		Description:     req.Description,
		CustomerKey:     req.CustomerKey,
		NotificationURL: req.WebhookURL,
		SuccessURL:      req.SuccessURL,
		FailURL:         req.FailURL,
	}

	resp, err := s.tbankClient.Init(innerReq)
	if err != nil {
		return nil, fmt.Errorf("tbank init payment: %w", err)
	}

	return &InitPaymentResponse{
		PaymentID:      resp.PaymentID,
		PaymentPageURL: resp.PaymentURL,
		OrderID:        resp.OrderID,
		Status:         resp.Status,
	}, nil
}

func (s *Service) GetPaymentStatus(ctx context.Context, paymentID string) (string, error) {
	req := &tinkoff.GetStateRequest{
		PaymentID: paymentID,
	}

	resp, err := s.tbankClient.GetState(req)
	if err != nil {
		return "", fmt.Errorf("tbank get payment status: %w", err)
	}

	return resp.Status, nil
}

// ParseNotification парсит и валидирует уведомление от Т-Банка
func (s *Service) ParseNotification(body []byte) (*tinkoff.Notification, error) {
	return s.tbankClient.ParseNotification(bytes.NewReader(body))
}
