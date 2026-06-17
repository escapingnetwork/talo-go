package talo

import (
	"context"
	"fmt"
	"net/http"
)

// RefundsService provides refund creation.
type RefundsService struct {
	client *Client
}

// Create creates a refund for a payment (FULL or PARTIAL).
func (s *RefundsService) Create(ctx context.Context, paymentID string, input CreateRefundRequest) (*Refund, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("paymentID is required")
	}
	var refund Refund
	path := fmt.Sprintf("/payments/%s/refunds", paymentID)
	err := s.client.doRequest(ctx, http.MethodPost, path, input, true, &refund)
	if err != nil {
		return nil, err
	}
	return &refund, nil
}
