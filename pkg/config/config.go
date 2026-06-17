package config

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

// Environment selects the Talo API base URL.
type Environment string

const (
	Production Environment = "production"
	Sandbox    Environment = "sandbox"
)

var defaultBaseURLs = map[Environment]string{
	Production: "https://api.talo.com.ar",
	Sandbox:    "https://sandbox-api.talo.com.ar",
}

// Config holds credentials and shared HTTP/token logic for Talo API clients.
type Config struct {
	ClientID     string
	ClientSecret string
	UserID       string
	Environment  Environment
	BaseURL      string
	HTTPClient   *http.Client // optional custom HTTP client

	// internal
	tokenManager *tokenManager
	httpClient   *http.Client
	baseURL      string
}

// Option is a functional option for configuring Config.
type Option func(*Config)

// WithEnvironment sets the environment (Production or Sandbox).
func WithEnvironment(env Environment) Option {
	return func(c *Config) { c.Environment = env }
}

// WithBaseURL overrides the base URL completely.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) { c.BaseURL = baseURL }
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Config) { c.HTTPClient = hc }
}

// New creates a new Config with the required credentials.
// It automatically sets up token management and HTTP client.
func New(clientID, clientSecret, userID string, opts ...Option) (*Config, error) {
	if clientID == "" || clientSecret == "" || userID == "" {
		return nil, errors.New("clientID, clientSecret and userID are required")
	}

	cfg := &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		UserID:       userID,
		Environment:  Production,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: defaultTimeout}
	}
	cfg.httpClient = cfg.HTTPClient

	base := cfg.BaseURL
	if base == "" {
		base = defaultBaseURLs[cfg.Environment]
		if base == "" {
			base = defaultBaseURLs[Production]
		}
	}
	cfg.baseURL = strings.TrimRight(base, "/")

	cfg.tokenManager = newTokenManager(cfg)

	return cfg, nil
}

type tokenManager struct {
	mu           sync.Mutex
	token        string
	expiresAt    time.Time
	baseURL      string
	clientID     string
	clientSecret string
	userID       string
	httpClient   *http.Client
}

func newTokenManager(cfg *Config) *tokenManager {
	return &tokenManager{
		baseURL:      cfg.baseURL,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		userID:       cfg.UserID,
		httpClient:   cfg.httpClient,
	}
}

func (tm *tokenManager) getAccessToken(ctx context.Context, forceRefresh bool) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !forceRefresh && tm.hasValidToken() {
		return tm.token, nil
	}

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

func extractJWTExpiration(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, errors.New("invalid jwt format")
	}
	payloadB64 := parts[1]
	payloadB64 = strings.ReplaceAll(payloadB64, "-", "+")
	payloadB64 = strings.ReplaceAll(payloadB64, "_", "/")
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

// DoRequest performs an API call with automatic authentication and 401 retry.
// This is used internally by resource clients.
func (c *Config) DoRequest(ctx context.Context, method, path string, body interface{}, authRequired bool, out interface{}) error {
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
		newToken, err := c.tokenManager.getAccessToken(ctx, true)
		if err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+newToken)
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
		resp = resp2
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, rawBody, requestID)
	}

	if out == nil {
		return nil
	}

	// Try envelope {data: ...} first, fallback to direct
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rawBody, &env); err == nil && len(env.Data) > 0 {
		return json.Unmarshal(env.Data, out)
	}
	return json.Unmarshal(rawBody, out)
}

func (c *Config) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	base := c.baseURL
	path = strings.TrimLeft(path, "/")
	return base + "/" + path
}

func parseAPIError(statusCode int, rawBody []byte, requestID string) error {
	var apiErr struct {
		Message string      `json:"message,omitempty"`
		Error   interface{} `json:"error,omitempty"`
		Detail  string      `json:"detail,omitempty"`
		Code    interface{} `json:"code,omitempty"`
		Errors  interface{} `json:"errors,omitempty"`
	}
	_ = json.Unmarshal(rawBody, &apiErr)

	msg := apiErr.Message
	if msg == "" {
		if s, ok := apiErr.Error.(string); ok && s != "" {
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

// TaloError represents an error returned by the Talo API.
type TaloError struct {
	StatusCode int         `json:"status_code,omitempty"`
	ErrorCode  interface{} `json:"error_code,omitempty"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	RawBody    string      `json:"raw_body,omitempty"`
}

func (e *TaloError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("talo: %s (status=%d, request_id=%s)", e.Message, e.StatusCode, e.RequestID)
	}
	return fmt.Sprintf("talo: %s (status=%d)", e.Message, e.StatusCode)
}
