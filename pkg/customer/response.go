package customer

import "encoding/json"

// Customer represents a customer/wallet.
type Customer struct {
	CustomerID        string          `json:"customer_id"`
	UserID            string          `json:"user_id,omitempty"`
	FullName          string          `json:"full_name,omitempty"`
	DocumentID        string          `json:"document_id,omitempty"`
	Email             string          `json:"email,omitempty"`
	Phone             string          `json:"phone,omitempty"`
	BankInfo          json.RawMessage `json:"bank_info,omitempty"`
	Balance           json.Number     `json:"balance,omitempty"`
	CreationTimestamp string          `json:"creation_timestamp,omitempty"`
	UpdateTimestamp   string          `json:"update_timestamp,omitempty"`
}

// Transaction represents a customer transaction.
type Transaction struct {
	TransactionID     string      `json:"transaction_id,omitempty"`
	PaymentID         string      `json:"payment_id,omitempty"`
	Status            string      `json:"status,omitempty"`
	Amount            json.Number `json:"amount,omitempty"`
	Currency          string      `json:"currency,omitempty"`
	CreationTimestamp string      `json:"creation_timestamp,omitempty"`
}
