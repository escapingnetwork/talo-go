package refund

import "encoding/json"

// Refund represents a created refund.
type Refund struct {
	RefundID  string      `json:"refund_id,omitempty"`
	PaymentID string      `json:"payment_id,omitempty"`
	Amount    json.Number `json:"amount,omitempty"`
	Currency  string      `json:"currency,omitempty"`
	Status    string      `json:"status,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
}
