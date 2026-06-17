package payment

import (
	"context"
	"fmt"
	"net/http"

	"github.com/escapingnetwork/talo-go/pkg/config"
)

// Client defines the interface for payment operations.
type Client interface {
	Create(ctx context.Context, req CreateRequest) (*Payment, error)
	Get(ctx context.Context, paymentID string) (*Payment, error)
	UpdateMetadata(ctx context.Context, paymentID string, req UpdateMetadataRequest) (*Payment, error)
}

// client is the concrete implementation.
type client struct {
	cfg *config.Config
}

// NewClient creates a new payment client.
func NewClient(cfg *config.Config) Client {
	return &client{cfg: cfg}
}

// Create creates a new payment.
func (c *client) Create(ctx context.Context, req CreateRequest) (*Payment, error) {
	var payment Payment
	err := c.cfg.DoRequest(ctx, http.MethodPost, "/payments/", req, true, &payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// Get retrieves a payment by ID.
func (c *client) Get(ctx context.Context, paymentID string) (*Payment, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("paymentID is required")
	}
	var payment Payment
	path := fmt.Sprintf("/payments/%s", paymentID)
	err := c.cfg.DoRequest(ctx, http.MethodGet, path, nil, true, &payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// UpdateMetadata updates payment metadata (currently only motive).
func (c *client) UpdateMetadata(ctx context.Context, paymentID string, req UpdateMetadataRequest) (*Payment, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("paymentID is required")
	}
	var payment Payment
	path := fmt.Sprintf("/payments/%s/metadata", paymentID)
	err := c.cfg.DoRequest(ctx, http.MethodPut, path, req, true, &payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
