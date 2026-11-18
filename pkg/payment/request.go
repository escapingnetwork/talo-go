package payment

import (
	"encoding/json"
	"fmt"
)

// StoreTypeCustom is the only store type accepted when creating a payment.
// A payment response may still carry another store_type when the API derived
// the store from a legacy integration.
const StoreTypeCustom = "custom"

// Store links a payment to a store that already exists in the Talo account.
// Payments created with a store use an ST- id prefix.
type Store struct {
	StoreType string `json:"store_type"`
	// StoreID is the store identifier in your system. Use a non-empty string or a number.
	StoreID any    `json:"store_id"`
	AppID   string `json:"app_id,omitempty"`
}

// CreatePaymentPrice represents the price for creating a payment.
type CreatePaymentPrice struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// ClientData contains optional payer information.
type ClientData struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	DNI       string `json:"dni,omitempty"`
	CUIT      string `json:"cuit,omitempty"`
}

// CreateRequest is the request body for creating a payment.
type CreateRequest struct {
	UserID         string             `json:"user_id"`
	Price          CreatePaymentPrice `json:"price"`
	PaymentOptions []string           `json:"payment_options"`
	ExternalID     string             `json:"external_id"`
	WebhookURL     string             `json:"webhook_url"`
	RedirectURL    string             `json:"redirect_url,omitempty"`
	Motive         string             `json:"motive,omitempty"`
	ClientData     *ClientData        `json:"client_data,omitempty"`
	PartnerID      string             `json:"partner_id,omitempty"`
	// Store associates the payment with an existing store. Only store_type "custom" is accepted.
	Store *Store `json:"store,omitempty"`
}

func validateStore(store *Store) error {
	if store == nil {
		return nil
	}
	if store.StoreType != StoreTypeCustom {
		return fmt.Errorf("store.store_type must be %q", StoreTypeCustom)
	}
	switch id := store.StoreID.(type) {
	case string:
		if id == "" {
			return fmt.Errorf("store.store_id is required")
		}
	case json.Number:
		if id == "" {
			return fmt.Errorf("store.store_id is required")
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return nil
	default:
		return fmt.Errorf("store.store_id must be a non-empty string or a number")
	}
	return nil
}

// UpdateMetadataRequest updates the motive of a payment.
type UpdateMetadataRequest struct {
	Motive string `json:"motive"`
}
