package talo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	refreshWindow  = 5 * time.Minute
	oneHour        = time.Hour
)

var (
	defaultBaseURLs = map[Environment]string{
		Production: "https://api.talo.com.ar",
		Sandbox:    "https://sandbox-api.talo.com.ar",
	}
)

// Client is the main entry point for the Talo Pay API.
type Client struct {
	cfg          *Config
	httpClient   *http.Client
	tokenManager *tokenManager
	baseURL      string

	// Sub-services for resource-oriented access (idiomatic Go)
	Payments  *PaymentsService
	Customers *CustomersService
	Partners  *PartnersService
	Refunds   *RefundsService
	Sandbox   *SandboxService
}

// tokenManager handles fetching and caching JWT access tokens.
type tokenManager struct {
	mu           sync.Mutex
	token        string
	expiresAt    time.Time
	baseURL      string
	clientID     string
	clientSecret string
	userID       string
	httpClient   *http.Client
	headers      http.Header
}

func newTokenManager(cfg *Config, httpClient *http.Client) *tokenManager {
	base := cfg.BaseURL
	if base == "" {
		base = defaultBaseURLs[cfg.Environment]
		if base == "" {
			base = defaultBaseURLs[Production]
		}
	}
	return &tokenManager{
		baseURL:      strings.TrimRight(base, "/"),
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		userID:       cfg.UserID,
		httpClient:   httpClient,
		headers:      cfg.Headers,
	}
}

func (tm *tokenManager) getAccessToken(ctx context.Context, forceRefresh bool) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !forceRefresh && tm.hasValidToken() {
		return tm.token, nil
	}

	// Simple in-memory singleflight: if already refreshing, wait? For simplicity we just proceed (rare contention)
	token, expiresAt, err := tm.fetchToken(ctx)
	if err != nil {
		return "", err
	}

	tm.token = token
	tm.expiresAt = expiresAt
	return token, nil
}

func (tm *tokenManager) hasValidToken() bool {
	if tm.token == "" || tm.expiresAt.IsZero() {
		return false
	}
	return time.Now().Before(tm.expiresAt.Add(-refreshWindow))
}

func (tm *tokenManager) fetchToken(ctx context.Context) (string, time.Time, error) {
	authURL := fmt.Sprintf("%s/users/%s/tokens", tm.baseURL, url.PathEscape(tm.userID))

	body := map[string]string{
		"client_id":     tm.clientID,
		"client_secret": tm.clientSecret,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if tm.headers != nil {
		for k, vv := range tm.headers {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, parseAPIError(resp.StatusCode, rawBody, resp.Header.Get("X-Request-ID"))
	}

	var env struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rawBody, &env); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse auth response: %w", err)
	}
	if env.Data.Token == "" {
		return "", time.Time{}, errors.New("empty token in auth response")
	}

	expiresAt := time.Now().Add(oneHour)
	if exp, err := extractJWTExpiration(env.Data.Token); err == nil && !exp.IsZero() {
		expiresAt = exp
	}

	return env.Data.Token, expiresAt, nil
}

// extractJWTExpiration parses the exp claim from a JWT (no signature validation).
func extractJWTExpiration(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, errors.New("invalid jwt format")
	}
	payloadB64 := parts[1]
	// base64url -> base64
	payloadB64 = strings.ReplaceAll(payloadB64, "-", "+")
	payloadB64 = strings.ReplaceAll(payloadB64, "_", "/")
	// padding
	switch len(payloadB64) % 4 {
	case 2:
		payloadB64 += "=="
	case 3:
		payloadB64 += "="
	}
	payload, err := base64.StdEncoding.DecodeString(payloadB64)
	if err != nil {
		return time.Time{}, err
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, err
	}
	if claims.Exp <= 0 {
		return time.Time{}, errors.New("no valid exp claim")
	}
	return time.Unix(claims.Exp, 0), nil
}

// NewClient creates a new Talo API client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.UserID == "" {
		return nil, errors.New("ClientID, ClientSecret and UserID are required")
	}

	if cfg.Environment == "" {
		cfg.Environment = Production
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	tm := newTokenManager(&cfg, httpClient)

	c := &Client{
		cfg:          &cfg,
		httpClient:   httpClient,
		tokenManager: tm,
		baseURL:      tm.baseURL,
	}

	// Initialize sub-services
	c.Payments = &PaymentsService{client: c}
	c.Customers = &CustomersService{client: c}
	c.Partners = &PartnersService{client: c}
	c.Refunds = &RefundsService{client: c}
	c.Sandbox = &SandboxService{client: c}

	return c, nil
}

// doRequest performs an authenticated (or not) HTTP request and handles token refresh on 401.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, authRequired bool, out interface{}) error {
	fullURL := c.buildURL(path)

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.Headers != nil {
		for k, vv := range c.cfg.Headers {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}

	var token string
	if authRequired {
		token, err = c.tokenManager.getAccessToken(ctx, false)
		if err != nil {
			return fmt.Errorf("get access token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	requestID := resp.Header.Get("X-Request-ID")

	if resp.StatusCode == http.StatusUnauthorized && authRequired {
		// Refresh token and retry once
		newToken, err := c.tokenManager.getAccessToken(ctx, true)
		if err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+newToken)

		// Re-create body reader if needed (since consumed)
		if body != nil {
			b, _ := json.Marshal(body)
			req.Body = io.NopCloser(bytes.NewReader(b))
		} else {
			req.Body = nil
		}

		resp2, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		rawBody, _ = io.ReadAll(resp2.Body)
		requestID = resp2.Header.Get("X-Request-ID")
		resp = resp2 // use for status check below
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, rawBody, requestID)
	}

	if out == nil {
		return nil
	}

	// Most responses are wrapped in { "data": ... }
	var env responseEnvelope
	if err := json.Unmarshal(rawBody, &env); err != nil {
		// Some endpoints (like sandbox faucet) may return unwrapped response
		if err := json.Unmarshal(rawBody, out); err != nil {
			return fmt.Errorf("unmarshal response: %w (raw: %s)", err, string(rawBody))
		}
		return nil
	}

	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	} else {
		// fallback direct unmarshal
		if err := json.Unmarshal(rawBody, out); err != nil {
			return fmt.Errorf("unmarshal response: %w (raw: %s)", err, string(rawBody))
		}
	}
	return nil
}

func (c *Client) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	base := strings.TrimRight(c.baseURL, "/")
	path = strings.TrimLeft(path, "/")
	return base + "/" + path
}

func parseAPIError(statusCode int, rawBody []byte, requestID string) error {
	var apiErr apiErrorBody
	_ = json.Unmarshal(rawBody, &apiErr)

	msg := apiErr.Message
	if msg == "" {
		if s, ok := apiErr.Error.(string); ok {
			msg = s
		} else if apiErr.Detail != "" {
			msg = apiErr.Detail
		} else {
			msg = fmt.Sprintf("HTTP %d", statusCode)
		}
	}

	return &TaloError{
		StatusCode: statusCode,
		ErrorCode:  apiErr.Code,
		Message:    msg,
		Details:    apiErr.Errors,
		RequestID:  requestID,
		RawBody:    string(rawBody),
	}
}

// --- Convenience top-level methods (aliases like in TS SDK) ---

func (c *Client) CreatePayment(ctx context.Context, input CreatePaymentRequest) (*Payment, error) {
	return c.Payments.Create(ctx, input)
}

func (c *Client) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	return c.Payments.Get(ctx, paymentID)
}

func (c *Client) UpdatePaymentMetadata(ctx context.Context, paymentID string, input UpdatePaymentMetadataRequest) (*Payment, error) {
	return c.Payments.UpdateMetadata(ctx, paymentID, input)
}

func (c *Client) CreateCustomer(ctx context.Context, input CreateCustomerRequest) (*Customer, error) {
	return c.Customers.Create(ctx, input)
}

func (c *Client) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	return c.Customers.Get(ctx, customerID)
}

func (c *Client) GetCustomerTransaction(ctx context.Context, customerID, transactionID string) (*CustomerTransaction, error) {
	return c.Customers.GetTransaction(ctx, customerID, transactionID)
}

func (c *Client) GetPartnerAuthorizationURL(partnerID string, referredUserID string) string {
	return c.Partners.GetAuthorizationURL(partnerID, referredUserID)
}

func (c *Client) ExchangePartnerToken(ctx context.Context, input PartnerTokenExchangeRequest) (*PartnerTokenExchangeResponse, error) {
	return c.Partners.ExchangeToken(ctx, input)
}

func (c *Client) GetPartnerAccount(ctx context.Context, userID string) (*PartnerAccount, error) {
	return c.Partners.GetAccount(ctx, userID)
}

func (c *Client) UpdatePartnerAccount(ctx context.Context, userID string, input UpdatePartnerAccountRequest) (*PartnerAccount, error) {
	return c.Partners.UpdateAccount(ctx, userID, input)
}

func (c *Client) CreateRefund(ctx context.Context, paymentID string, input CreateRefundRequest) (*Refund, error) {
	return c.Refunds.Create(ctx, paymentID, input)
}

func (c *Client) SimulateCvuTransfer(ctx context.Context, cvu string, input SimulateFaucetRequest) (*SimulateFaucetResponse, error) {
	return c.Sandbox.SimulateCvuTransfer(ctx, cvu, input)
}
