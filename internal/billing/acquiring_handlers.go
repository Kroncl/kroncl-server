package billing

import (
	"io"
	"kroncl-server/internal/core"
	"kroncl-server/internal/pricing"
	"log"
	"net/http"
)

// WebhookHandler обрабатывает уведомления от Т-Банка
func (h *Handlers) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Failed to read body: %v", err)
		core.SendValidationError(w, "Failed to read body")
		return
	}
	defer r.Body.Close()

	// Парсим и валидируем уведомление
	notification, err := h.service.ParseNotification(body)
	if err != nil {
		log.Printf("❌ T-Bank notification error: %v", err)
		core.SendValidationError(w, "Invalid notification")
		return
	}

	log.Printf("📩 T-Bank notification: OrderID=%s, Status=%s, PaymentID=%s",
		notification.OrderID, notification.Status, notification.PaymentID)

	// Определяем статус транзакции
	var txStatus pricing.TransactionStatus
	switch notification.Status {
	case "CONFIRMED", "AUTHORIZED":
		txStatus = pricing.TransactionStatusSuccess
	case "REJECTED", "CANCELED", "DEADLINE_EXPIRED":
		txStatus = pricing.TransactionStatusUnsuccess
	default:
		log.Printf("⚠️ Unknown status from T-Bank: %s", notification.Status)
		txStatus = pricing.TransactionStatusPending
	}

	// Обновляем статус транзакции в БД
	_, err = h.service.pricingService.UpdateTransactionStatus(r.Context(), notification.OrderID, txStatus)
	if err != nil {
		log.Printf("❌ Failed to update transaction status: %v", err)
		core.SendInternalError(w, "Failed to update transaction")
		return
	}

	// Возвращаем успешный ответ банку
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success"}`))
}
