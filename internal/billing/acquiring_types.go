package billing

type InitPaymentRequest struct {
	OrderID     string
	Amount      uint64 // в копейках
	Description string
	CustomerKey string
	WebhookURL  string
	SuccessURL  string
	FailURL     string
}

type InitPaymentResponse struct {
	PaymentID      string
	PaymentPageURL string
	OrderID        string
	Status         string
}
