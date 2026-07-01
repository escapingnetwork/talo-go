package partner

import (
	"testing"

	"github.com/escapingnetwork/talo-go/pkg/config"
)

func newTestConfig(t *testing.T, opts ...config.Option) *config.Config {
	t.Helper()
	cfg, err := config.New("client-id", "client-secret", "user-id", opts...)
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	return cfg
}

func TestGetAuthorizationURL_Sandbox(t *testing.T) {
	client := NewClient(newTestConfig(t, config.WithEnvironment(config.Sandbox)))

	got := client.GetAuthorizationURL("partner-123", "biz-uuid")
	want := "https://sandbox.talo.com.ar/authorize/partner-123?referred_user_id=biz-uuid"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGetAuthorizationURL_Production(t *testing.T) {
	client := NewClient(newTestConfig(t))

	got := client.GetAuthorizationURL("partner-123", "")
	want := "https://app.talo.com.ar/authorize/partner-123"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGetAuthorizationURL_EmptyPartnerID(t *testing.T) {
	client := NewClient(newTestConfig(t, config.WithEnvironment(config.Sandbox)))

	if got := client.GetAuthorizationURL("", "biz-uuid"); got != "" {
		t.Fatalf("got %q, want empty string", got)
	}
}

func TestGetAuthorizationURL_QueryEscapesReferredUserID(t *testing.T) {
	client := NewClient(newTestConfig(t))

	got := client.GetAuthorizationURL("partner-123", "biz uuid&foo=bar")
	want := "https://app.talo.com.ar/authorize/partner-123?referred_user_id=biz+uuid%26foo%3Dbar"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGetAuthorizationURL_CustomAuthorizeBaseURL(t *testing.T) {
	client := NewClient(newTestConfig(t, config.WithAuthorizeBaseURL("https://custom.example/authorize")))

	got := client.GetAuthorizationURL("partner-123", "")
	want := "https://custom.example/authorize/partner-123"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}