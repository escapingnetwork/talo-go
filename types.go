package talo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Environment represents the Talo API environment.
type Environment string

const (
	// Production is the live environment: https://api.talo.com.ar
	Production Environment = "production"
	// Sandbox is the test environment: https://sandbox-api.talo.com.ar
	Sandbox Environment = "sandbox"
)

// Config holds the configuration for the Talo client.
type Config struct {
	// ClientID is your Talo client identifier.
	ClientID string
	// ClientSecret is your Talo client secret.
	ClientSecret string
	// UserID is the Talo user ID associated with the credentials.
	UserID string
	// Environment selects the base URL. Defaults to Production.
	// Ignored if BaseURL is set.
	Environment Environment
	// BaseURL overrides the environment base URL (e.g. for custom deployments).
	BaseURL string
	// HTTPClient allows providing a custom *http.Client (e.g. with timeouts, transport).
	// If nil, a default client with reasonable timeouts is used.
	HTTPClient *http.Client
	// Headers are additional headers to include in every request (e.g. User-Agent).
	Headers http.Header
}

// ClientData represents optional client information for payments.
type ClientData struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	DNI       string `json:"dni,omitempty"`
	CUIT      string `json:"cuit,omitempty"`
}

// CreatePaymentPrice is the price object for creating payments (amount in whole ARS units).
type CreatePaymentPrice struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// CreatePaymentRequest is the input for creating a new payment.
type CreatePaymentRequest struct {
	UserID         string                 `json:"user_id"`
	Price          CreatePaymentPrice     `json:"price"`
	PaymentOptions []string               `json:"payment_options"` // typically ["transfer"]
	ExternalID     string                 `json:"external_id"`
	WebhookURL     string                 `json:"webhook_url"`
	RedirectURL    string                 `json:"redirect_url,omitempty"`
	Motive         string                 `json:"motive,omitempty"`
	ClientData     *ClientData            `json:"client_data,omitempty"`
	PartnerID      string                 `json:"partner_id,omitempty"`
}

// UpdatePaymentMetadataRequest updates the motive on a payment.
type UpdatePaymentMetadataRequest struct {
	Motive string `json:"motive"`
}

// Payment represents a payment object returned by the API.
type Payment struct {
	ID                    string                 `json:"id"`
	PaymentStatus         string                 `json:"payment_status"`
	UserID                string                 `json:"user_id,omitempty"`
	Quotes                json.RawMessage        `json:"quotes,omitempty"` // flexible
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

// Price represents monetary amount.
type Price struct {
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
}

// CreateCustomerRequest registers a customer/wallet.
type CreateCustomerRequest struct {
	UserID     string `json:"user_id"`
	FullName   string `json:"full_name"`
	DocumentID string `json:"document_id"`
	Email      string `json:"email"`
	Phone      string `json:"phone,omitempty"`
	CVU        string `json:"cvu,omitempty"`
	CBU        string `json:"cbu,omitempty"`
	Alias      string `json:"alias,omitempty"`
}

// Customer represents a customer object.
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

// CustomerTransaction represents a transaction for a customer.
type CustomerTransaction struct {
	TransactionID     string      `json:"transaction_id,omitempty"`
	PaymentID         string      `json:"payment_id,omitempty"`
	Status            string      `json:"status,omitempty"`
	Amount            json.Number `json:"amount,omitempty"`
	Currency          string      `json:"currency,omitempty"`
	CreationTimestamp string      `json:"creation_timestamp,omitempty"`
}

// CreateRefundRequest for creating refunds. Use RefundType FULL or PARTIAL.
type CreateRefundRequest struct {
	RefundType string      `json:"refund_type"` // "FULL" or "PARTIAL"
	Amount     string      `json:"amount,omitempty"`
	Currency   string      `json:"currency,omitempty"`
	Blame      RefundBlame `json:"blame"`
	UserID     string      `json:"user_id"`
}

// RefundBlame identifies who initiated the refund.
type RefundBlame struct {
	TeamID string `json:"team_id"`
	Mail   string `json:"mail"`
}

// Refund represents a refund response.
type Refund struct {
	RefundID string      `json:"refund_id,omitempty"`
	PaymentID string     `json:"payment_id,omitempty"`
	Amount   json.Number `json:"amount,omitempty"`
	Currency string      `json:"currency,omitempty"`
	Status   string      `json:"status,omitempty"`
	CreatedAt string     `json:"created_at,omitempty"`
}

// PartnerTokenExchangeRequest exchanges OAuth-like code for tokens (for partners).
type PartnerTokenExchangeRequest struct {
	Code        string `json:"code"`
	ClientID    string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// PartnerTokenExchangeResponse is the result of token exchange.
type PartnerTokenExchangeResponse struct {
	Token         string `json:"token"`
	UserID        string `json:"user_id"`
	ReferredUserID string `json:"referred_user_id,omitempty"`
}

// PartnerAccount represents partner account config.
type PartnerAccount struct {
	AccountStatus     string          `json:"account_status,omitempty"`
	AliasPrefix       string          `json:"alias_prefix,omitempty"`
	CancellationPeriod int            `json:"cancellation_period,omitempty"`
	TransferTolerance int             `json:"transfer_tolerance,omitempty"`
	PayoutSchedule    json.RawMessage `json:"payout_schedule,omitempty"`
	UserID            string          `json:"user_id,omitempty"`
	PartnerConfig     json.RawMessage `json:"partner_config,omitempty"`
}

// UpdatePartnerAccountRequest updates partner settings. At least one field required.
type UpdatePartnerAccountRequest struct {
	AliasPrefix       string          `json:"alias_prefix,omitempty"`
	CancellationPeriod *int           `json:"cancellation_period,omitempty"`
	TransferTolerance *int            `json:"transfer_tolerance,omitempty"`
	PayoutSchedule    json.RawMessage `json:"payout_schedule,omitempty"`
}

// SimulateFaucetRequest for sandbox CVU top-up simulation.
type SimulateFaucetRequest struct {
	Amount json.Number `json:"amount"`
}

// SimulateFaucetResponse from sandbox faucet.
type SimulateFaucetResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// TaloError is the error type returned for API failures.
type TaloError struct {
	StatusCode int         `json:"status_code,omitempty"`
	ErrorCode  interface{} `json:"error_code,omitempty"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	RawBody    string      `json:"raw_body,omitempty"`
}

func (e *TaloError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("talo: %s (status=%d, request_id=%s)", e.Message, e.StatusCode, e.RequestID)
	}
	return fmt.Sprintf("talo: %s (status=%d)", e.Message, e.StatusCode)
}

// responseEnvelope is a generic envelope used by most Talo responses.
type responseEnvelope struct {
	Message string          `json:"message,omitempty"`
	Status  string          `json:"status,omitempty"`
	Error   bool            `json:"error,omitempty"`
	Code    int             `json:"code,omitempty"`
	Data    json.RawMessage `json:"data"`
}

// apiErrorBody for parsing error responses.
type apiErrorBody struct {
	Message string      `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Detail  string      `json:"detail,omitempty"`
	Code    interface{} `json:"code,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}
