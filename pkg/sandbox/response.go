package sandbox

// SimulateFaucetResponse from the sandbox faucet endpoint.
type SimulateFaucetResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}
