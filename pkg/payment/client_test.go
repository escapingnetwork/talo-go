package payment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/escapingnetwork/talo-go/pkg/config"
)

func testConfig(t *testing.T, baseURL string) *config.Config {
	t.Helper()
	cfg, err := config.New("client-id", "client-secret", "user-id", config.WithBaseURL(baseURL))
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	return cfg
}

func TestCreateForwardsStore(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/tokens"):
			io.WriteString(w, `{"data":{"token":"token_abc"}}`)
		case r.URL.Path == "/payments/":
			body, _ = io.ReadAll(r.Body)
			io.WriteString(w, `{"data":{"id":"ST-user-id-order_4","payment_status":"PENDING","store":{"store_type":"custom","store_id":"tienda-centro"}}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	payment, err := NewClient(testConfig(t, srv.URL)).Create(context.Background(), CreateRequest{
		UserID:         "user-id",
		Price:          CreatePaymentPrice{Amount: 10000, Currency: "ARS"},
		PaymentOptions: []string{"transfer"},
		ExternalID:     "order_4",
		WebhookURL:     "https://example.com/webhooks/talo",
		Store: &Store{
			StoreType: StoreTypeCustom,
			StoreID:   "tienda-centro",
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var sent struct {
		Store *Store `json:"store"`
	}
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if sent.Store == nil || sent.Store.StoreType != StoreTypeCustom || sent.Store.StoreID != "tienda-centro" {
		t.Fatalf("sent store = %#v", sent.Store)
	}
	if payment.Store == nil || payment.Store.StoreType != "custom" || payment.Store.StoreID != "tienda-centro" {
		t.Fatalf("response store = %#v", payment.Store)
	}
	if payment.ID != "ST-user-id-order_4" {
		t.Fatalf("id = %q", payment.ID)
	}
}

func TestCreateMarshalsNumericStoreID(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/tokens") {
			io.WriteString(w, `{"data":{"token":"token_abc"}}`)
			return
		}
		body, _ = io.ReadAll(r.Body)
		io.WriteString(w, `{"data":{"id":"ST-1","payment_status":"PENDING","store":{"store_type":"pow","store_id":42}}}`)
	}))
	defer srv.Close()

	payment, err := NewClient(testConfig(t, srv.URL)).Create(context.Background(), CreateRequest{
		ExternalID: "order_6",
		WebhookURL: "https://example.com/webhooks/talo",
		Store:      &Store{StoreType: StoreTypeCustom, StoreID: 42, AppID: "app-1"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var sent map[string]any
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	store, _ := sent["store"].(map[string]any)
	if store["store_id"] != float64(42) || store["app_id"] != "app-1" {
		t.Fatalf("sent store = %#v", store)
	}
	if payment.Store == nil || payment.Store.StoreType != "pow" || payment.Store.StoreID != float64(42) {
		t.Fatalf("response store = %#v", payment.Store)
	}
}

func TestCreateRejectsUnsupportedStoreType(t *testing.T) {
	calledPayments := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/payments") {
			calledPayments = true
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := NewClient(testConfig(t, srv.URL)).Create(context.Background(), CreateRequest{
		Store: &Store{StoreType: "tiendanube", StoreID: "tienda-centro"},
	})
	if err == nil || !strings.Contains(err.Error(), `store.store_type must be "custom"`) {
		t.Fatalf("err = %v", err)
	}
	if calledPayments {
		t.Fatal("Create called /payments/ for an invalid store")
	}
}

func TestCreateRejectsEmptyStoreID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s", r.URL.Path)
	}))
	defer srv.Close()

	_, err := NewClient(testConfig(t, srv.URL)).Create(context.Background(), CreateRequest{
		Store: &Store{StoreType: StoreTypeCustom, StoreID: ""},
	})
	if err == nil || !strings.Contains(err.Error(), "store.store_id is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateAndGetAlwaysSendTaloToken(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path+" auth="+r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/tokens") {
			io.WriteString(w, `{"data":{"token":"token_abc"}}`)
			return
		}
		io.WriteString(w, `{"data":{"id":"payment_123","payment_status":"PENDING"}}`)
	}))
	defer srv.Close()

	client := NewClient(testConfig(t, srv.URL))
	if _, err := client.Create(context.Background(), CreateRequest{ExternalID: "order_1"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := client.Get(context.Background(), "payment_123"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	want := []string{
		"POST /users/user-id/tokens auth=",
		"POST /payments/ auth=Bearer token_abc",
		"GET /payments/payment_123 auth=Bearer token_abc",
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("requests:\n%s", strings.Join(seen, "\n"))
	}
}

func TestRetryStillSendsRefreshedToken(t *testing.T) {
	var paymentAuth []string
	tokenCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/tokens") {
			tokenCalls++
			token := "token_abc"
			if tokenCalls > 1 {
				token = "token_refreshed"
			}
			io.WriteString(w, `{"data":{"token":"`+token+`"}}`)
			return
		}
		paymentAuth = append(paymentAuth, r.Header.Get("Authorization"))
		if len(paymentAuth) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, `{"message":"unauthorized"}`)
			return
		}
		io.WriteString(w, `{"data":{"id":"payment_123","payment_status":"PENDING"}}`)
	}))
	defer srv.Close()

	if _, err := NewClient(testConfig(t, srv.URL)).Get(context.Background(), "payment_123"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := []string{"Bearer token_abc", "Bearer token_refreshed"}
	if strings.Join(paymentAuth, ",") != strings.Join(want, ",") {
		t.Fatalf("payment auth = %#v", paymentAuth)
	}
}
