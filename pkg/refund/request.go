package refund

// Blame identifies the team/user who requested the refund.
type Blame struct {
	TeamID string `json:"team_id"`
	Mail   string `json:"mail"`
}

// CreateRequest creates a refund for a payment.
type CreateRequest struct {
	RefundType string `json:"refund_type"` // "FULL" or "PARTIAL"
	Amount     string `json:"amount,omitempty"`
	Currency   string `json:"currency,omitempty"`
	Blame      Blame  `json:"blame"`
	UserID     string `json:"user_id"`
}
