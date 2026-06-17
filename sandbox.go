package talo

import (
	"context"
	"fmt"
	"net/http"
)

// SandboxService provides sandbox-only utilities like faucet simulation.
type SandboxService struct {
	client *Client
}

// SimulateCvuTransfer simulates an incoming transfer to a CVU (sandbox only).
func (s *SandboxService) SimulateCvuTransfer(ctx context.Context, cvu string, input SimulateFaucetRequest) (*SimulateFaucetResponse, error) {
	if cvu == "" {
		return nil, fmt.Errorf("cvu is required")
	}
	var resp SimulateFaucetResponse
	path := fmt.Sprintf("/cvu/%s/faucet", cvu)
	// Note: in original TS, this one returns the response directly (not .data), our doRequest handles both
	err := s.client.doRequest(ctx, http.MethodPost, path, input, true, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
