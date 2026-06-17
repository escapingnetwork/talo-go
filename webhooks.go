package talo

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// WebhookEvent is a minimal representation of Talo webhook payloads (payment.updated).
// Extend as needed based on actual payload shape.
type WebhookEvent struct {
	Type       string `json:"type"`
	PaymentID  string `json:"payment_id"`
	ExternalID string `json:"external_id,omitempty"`
	Status     string `json:"status,omitempty"`
	// Add other fields from actual webhook if known
	Raw json.RawMessage `json:"-"`
}

// ParseWebhook parses and validates a webhook request body.
// It does not perform signature verification (if Talo uses any, add it here).
// Returns the parsed event or error.
func ParseWebhook(rawBody []byte) (*WebhookEvent, error) {
	if len(rawBody) == 0 {
		return nil, errors.New("empty webhook body")
	}
	var event WebhookEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		return nil, errors.New("invalid webhook JSON payload")
	}
	event.Raw = rawBody
	if event.Type == "" || event.PaymentID == "" {
		return nil, errors.New("webhook payload missing required fields (type or payment_id)")
	}
	return &event, nil
}

// WebhookHandler is a simple http.HandlerFunc example for handling Talo webhooks.
// It parses the body, optionally fetches latest payment state using the client,
// and calls the provided handler func.
func (c *Client) WebhookHandler(onPaymentUpdated func(event *WebhookEvent, payment *Payment, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := readBody(r)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		event, err := ParseWebhook(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var payment *Payment
		if c != nil && event.PaymentID != "" {
			// Best-effort enrich with latest payment state (like TS SDK does)
			payment, _ = c.GetPayment(r.Context(), event.PaymentID)
		}

		if onPaymentUpdated != nil {
			if err := onPaymentUpdated(event, payment, r); err != nil {
				http.Error(w, "webhook handler failed", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"received":       true,
			"payment_id":     event.PaymentID,
			"payment_status": event.Status,
		})
	}
}

func readBody(r *http.Request) ([]byte, error) {
	// Simple read; in production consider http.MaxBytesReader etc.
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
