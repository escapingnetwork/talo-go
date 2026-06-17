package customer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/escapingnetwork/talo-go/pkg/config"
)

// Client defines customer operations.
type Client interface {
	Create(ctx context.Context, req CreateRequest) (*Customer, error)
	Get(ctx context.Context, customerID string) (*Customer, error)
	GetTransaction(ctx context.Context, customerID, transactionID string) (*Transaction, error)
}

type client struct {
	cfg *config.Config
}

// NewClient creates a new customer client.
func NewClient(cfg *config.Config) Client {
	return &client{cfg: cfg}
}

func (c *client) Create(ctx context.Context, req CreateRequest) (*Customer, error) {
	var customer Customer
	err := c.cfg.DoRequest(ctx, http.MethodPost, "/customers/", req, true, &customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (c *client) Get(ctx context.Context, customerID string) (*Customer, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customerID is required")
	}
	var customer Customer
	path := fmt.Sprintf("/customers/%s", customerID)
	err := c.cfg.DoRequest(ctx, http.MethodGet, path, nil, true, &customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (c *client) GetTransaction(ctx context.Context, customerID, transactionID string) (*Transaction, error) {
	if customerID == "" || transactionID == "" {
		return nil, fmt.Errorf("customerID and transactionID are required")
	}
	var tx Transaction
	path := fmt.Sprintf("/customers/%s/transactions/%s", customerID, transactionID)
	err := c.cfg.DoRequest(ctx, http.MethodGet, path, nil, true, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
