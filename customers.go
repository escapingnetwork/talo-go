package talo

import (
	"context"
	"fmt"
	"net/http"
)

// CustomersService provides methods for customer/wallet management.
type CustomersService struct {
	client *Client
}

// Create registers a new customer (wallet) to receive transfers.
func (s *CustomersService) Create(ctx context.Context, input CreateCustomerRequest) (*Customer, error) {
	var customer Customer
	err := s.client.doRequest(ctx, http.MethodPost, "/customers/", input, true, &customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// Get retrieves a customer by ID.
func (s *CustomersService) Get(ctx context.Context, customerID string) (*Customer, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customerID is required")
	}
	var customer Customer
	path := fmt.Sprintf("/customers/%s", customerID)
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, true, &customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// GetTransaction retrieves a specific incoming transaction for a customer.
func (s *CustomersService) GetTransaction(ctx context.Context, customerID, transactionID string) (*CustomerTransaction, error) {
	if customerID == "" || transactionID == "" {
		return nil, fmt.Errorf("customerID and transactionID are required")
	}
	var tx CustomerTransaction
	path := fmt.Sprintf("/customers/%s/transactions/%s", customerID, transactionID)
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, true, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
