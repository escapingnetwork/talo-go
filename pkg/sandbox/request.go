package sandbox

import "encoding/json"

// SimulateFaucetRequest is used to simulate a transfer in sandbox.
type SimulateFaucetRequest struct {
	Amount json.Number `json:"amount"`
}
