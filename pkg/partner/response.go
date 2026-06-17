package partner

import "encoding/json"

// TokenExchangeResponse is returned after exchanging a partner code.
type TokenExchangeResponse struct {
	Token          string `json:"token"`
	UserID         string `json:"user_id"`
	ReferredUserID string `json:"referred_user_id,omitempty"`
}

// Account represents partner account configuration.
type Account struct {
	AccountStatus      string          `json:"account_status,omitempty"`
	AliasPrefix        string          `json:"alias_prefix,omitempty"`
	CancellationPeriod int             `json:"cancellation_period,omitempty"`
	TransferTolerance  int             `json:"transfer_tolerance,omitempty"`
	PayoutSchedule     json.RawMessage `json:"payout_schedule,omitempty"`
	UserID             string          `json:"user_id,omitempty"`
	PartnerConfig      json.RawMessage `json:"partner_config,omitempty"`
}
