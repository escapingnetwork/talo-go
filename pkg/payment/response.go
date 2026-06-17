package payment

import "encoding/json"

// Payment represents a payment returned by the API.
type Payment struct {
	ID                    string                 `json:"id"`
	PaymentStatus         string                 `json:"payment_status"`
	UserID                string                 `json:"user_id,omitempty"`
	Quotes                json.RawMessage        `json:"quotes,omitempty"`
	TransactionFields     json.RawMessage        `json:"transaction_fields,omitempty"`
	Transactions          json.RawMessage        `json:"transactions,omitempty"`
	PaymentURL            string                 `json:"payment_url,omitempty"`
	ExternalID            string                 `json:"external_id,omitempty"`
	ExpirationTimestamp   string                 `json:"expiration_timestamp,omitempty"`
	CreationTimestamp     string                 `json:"creation_timestamp,omitempty"`
	LastModifiedTimestamp string                 `json:"last_modified_timestamp,omitempty"`
	Price                 *Price                 `json:"price,omitempty"`
	PaymentOptions        []string               `json:"payment_options,omitempty"`
	WebhookURL            string                 `json:"webhook_url,omitempty"`
	RedirectURL           string                 `json:"redirect_url,omitempty"`
	Motive                string                 `json:"motive,omitempty"`
	ClientData            *ClientData            `json:"client_data,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	PartnerID             string                 `json:"partner_id,omitempty"`
}

// Price represents a monetary amount.
type Price struct {
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
}
