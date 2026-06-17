package partner

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/escapingnetwork/talo-go/pkg/config"
)

const partnerAuthorizeBase = "https://app.talo.com.ar/authorize"

// Client defines partner operations.
type Client interface {
	GetAuthorizationURL(partnerID, referredUserID string) string
	ExchangeToken(ctx context.Context, req TokenExchangeRequest) (*TokenExchangeResponse, error)
	GetAccount(ctx context.Context, userID string) (*Account, error)
	UpdateAccount(ctx context.Context, userID string, req UpdateAccountRequest) (*Account, error)
}

type client struct {
	cfg *config.Config
}

// NewClient creates a new partner client.
func NewClient(cfg *config.Config) Client {
	return &client{cfg: cfg}
}

// GetAuthorizationURL builds the redirect URL for partner onboarding.
func (c *client) GetAuthorizationURL(partnerID, referredUserID string) string {
	if partnerID == "" {
		return ""
	}
	u := fmt.Sprintf("%s/%s", partnerAuthorizeBase, url.PathEscape(partnerID))
	if referredUserID != "" {
		u += "?referred_user_id=" + url.QueryEscape(referredUserID)
	}
	return u
}

func (c *client) ExchangeToken(ctx context.Context, req TokenExchangeRequest) (*TokenExchangeResponse, error) {
	var resp TokenExchangeResponse
	// This endpoint does not require the user access token
	err := c.cfg.DoRequest(ctx, http.MethodPost, "/auth/tokens", req, false, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) GetAccount(ctx context.Context, userID string) (*Account, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	var account Account
	path := fmt.Sprintf("/users/%s/account", userID)
	err := c.cfg.DoRequest(ctx, http.MethodGet, path, nil, true, &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (c *client) UpdateAccount(ctx context.Context, userID string, req UpdateAccountRequest) (*Account, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	var account Account
	path := fmt.Sprintf("/users/%s", userID)
	err := c.cfg.DoRequest(ctx, http.MethodPatch, path, req, true, &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}
