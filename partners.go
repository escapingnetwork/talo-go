package talo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// PartnersService provides partner onboarding and account management methods.
type PartnersService struct {
	client *Client
}

const partnerAuthorizeBase = "https://app.talo.com.ar/authorize"

// GetAuthorizationURL builds the URL to redirect users for partner onboarding.
// referredUserID is optional.
func (s *PartnersService) GetAuthorizationURL(partnerID, referredUserID string) string {
	if partnerID == "" {
		return ""
	}
	u := fmt.Sprintf("%s/%s", partnerAuthorizeBase, url.PathEscape(partnerID))
	if referredUserID != "" {
		u += "?referred_user_id=" + url.QueryEscape(referredUserID)
	}
	return u
}

// ExchangeToken exchanges the authorization code (from redirect) for a partner token and user mapping.
// Note: this call does NOT require the main user access token.
func (s *PartnersService) ExchangeToken(ctx context.Context, input PartnerTokenExchangeRequest) (*PartnerTokenExchangeResponse, error) {
	var resp PartnerTokenExchangeResponse
	// authRequired = false for this endpoint
	err := s.client.doRequest(ctx, http.MethodPost, "/auth/tokens", input, false, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAccount fetches the partner account configuration for a user.
func (s *PartnersService) GetAccount(ctx context.Context, userID string) (*PartnerAccount, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	var account PartnerAccount
	path := fmt.Sprintf("/users/%s/account", userID)
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, true, &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// UpdateAccount updates partner account settings (alias_prefix, tolerances, payout schedule etc).
func (s *PartnersService) UpdateAccount(ctx context.Context, userID string, input UpdatePartnerAccountRequest) (*PartnerAccount, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	var account PartnerAccount
	path := fmt.Sprintf("/users/%s", userID)
	err := s.client.doRequest(ctx, http.MethodPatch, path, input, true, &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}
