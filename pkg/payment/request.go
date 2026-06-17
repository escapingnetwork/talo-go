package payment

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
	UserID         string                 `json:"user_id"`
	Price          CreatePaymentPrice     `json:"price"`
	PaymentOptions []string               `json:"payment_options"`
	ExternalID     string                 `json:"external_id"`
	WebhookURL     string                 `json:"webhook_url"`
	RedirectURL    string                 `json:"redirect_url,omitempty"`
	Motive         string                 `json:"motive,omitempty"`
	ClientData     *ClientData            `json:"client_data,omitempty"`
	PartnerID      string                 `json:"partner_id,omitempty"`
}

// UpdateMetadataRequest updates the motive of a payment.
type UpdateMetadataRequest struct {
	Motive string `json:"motive"`
}
