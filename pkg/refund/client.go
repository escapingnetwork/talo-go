package refund

import (
	"context"
	"fmt"
	"net/http"

	"github.com/escapingnetwork/talo-go/pkg/config"
)

// Client defines refund operations.
type Client interface {
	Create(ctx context.Context, paymentID string, req CreateRequest) (*Refund, error)
}

type client struct {
	cfg *config.Config
}

// NewClient creates a new refund client.
func NewClient(cfg *config.Config) Client {
	return &client{cfg: cfg}
}

func (c *client) Create(ctx context.Context, paymentID string, req CreateRequest) (*Refund, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("paymentID is required")
	}
	var refund Refund
	path := fmt.Sprintf("/payments/%s/refunds", paymentID)
	err := c.cfg.DoRequest(ctx, http.MethodPost, path, req, true, &refund)
	if err != nil {
		return nil, err
	}
	return &refund, nil
}
