package partner

import "encoding/json"

// TokenExchangeRequest exchanges an authorization code for tokens.
type TokenExchangeRequest struct {
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// UpdateAccountRequest updates partner account settings.
type UpdateAccountRequest struct {
	AliasPrefix        string          `json:"alias_prefix,omitempty"`
	CancellationPeriod *int            `json:"cancellation_period,omitempty"`
	TransferTolerance  *int            `json:"transfer_tolerance,omitempty"`
	PayoutSchedule     json.RawMessage `json:"payout_schedule,omitempty"`
}
