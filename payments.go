package talo

import (
	"context"
	"fmt"
	"net/http"
)

// PaymentsService provides methods for payments.
type PaymentsService struct {
	client *Client
}

// Create creates a new payment and returns transfer instructions / payment details.
func (s *PaymentsService) Create(ctx context.Context, input CreatePaymentRequest) (*Payment, error) {
	var payment Payment
	err := s.client.doRequest(ctx, http.MethodPost, "/payments/", input, true, &payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// Get retrieves a payment by its ID.
func (s *PaymentsService) Get(ctx context.Context, paymentID string) (*Payment, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("paymentID is required")
	}
	var payment Payment
	path := fmt.Sprintf("/payments/%s", paymentID)
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, true, &payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// UpdateMetadata updates the motive/description of a payment.
func (s *PaymentsService) UpdateMetadata(ctx context.Context, paymentID string, input UpdatePaymentMetadataRequest) (*Payment, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("paymentID is required")
	}
	var payment Payment
	path := fmt.Sprintf("/payments/%s/metadata", paymentID)
	err := s.client.doRequest(ctx, http.MethodPut, path, input, true, &payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
