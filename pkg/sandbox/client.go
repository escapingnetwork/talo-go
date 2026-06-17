package sandbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/escapingnetwork/talo-go/pkg/config"
)

// Client defines sandbox utilities.
type Client interface {
	SimulateCvuTransfer(ctx context.Context, cvu string, req SimulateFaucetRequest) (*SimulateFaucetResponse, error)
}

type client struct {
	cfg *config.Config
}

// NewClient creates a new sandbox client.
func NewClient(cfg *config.Config) Client {
	return &client{cfg: cfg}
}

func (c *client) SimulateCvuTransfer(ctx context.Context, cvu string, req SimulateFaucetRequest) (*SimulateFaucetResponse, error) {
	if cvu == "" {
		return nil, fmt.Errorf("cvu is required")
	}
	var resp SimulateFaucetResponse
	path := fmt.Sprintf("/cvu/%s/faucet", cvu)
	err := c.cfg.DoRequest(ctx, http.MethodPost, path, req, true, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
