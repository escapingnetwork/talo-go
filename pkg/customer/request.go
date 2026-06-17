package customer

// CreateRequest registers a new customer/wallet.
type CreateRequest struct {
	UserID     string `json:"user_id"`
	FullName   string `json:"full_name"`
	DocumentID string `json:"document_id"`
	Email      string `json:"email"`
	Phone      string `json:"phone,omitempty"`
	CVU        string `json:"cvu,omitempty"`
	CBU        string `json:"cbu,omitempty"`
	Alias      string `json:"alias,omitempty"`
}
