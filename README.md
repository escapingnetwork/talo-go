# Talo Pay SDK (Go)

Type-safe Go client for the Talo Transfers API.

## Install

```bash
go get github.com/escapingnetwork/talo-go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/escapingnetwork/talo-go"
)

func main() {
	client, err := talo.NewClient(talo.Config{
		ClientID:     os.Getenv("TALO_CLIENT_ID"),
		ClientSecret: os.Getenv("TALO_CLIENT_SECRET"),
		UserID:       os.Getenv("TALO_USER_ID"),
		Environment:  talo.Sandbox, // or talo.Production
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create a payment
	payment, err := client.CreatePayment(ctx, talo.CreatePaymentRequest{
		UserID: os.Getenv("TALO_USER_ID"),
		Price: talo.CreatePaymentPrice{
			Amount:   1500, // whole ARS units
			Currency: "ARS",
		},
		PaymentOptions: []string{"transfer"},
		ExternalID:     "order_12345",
		WebhookURL:     "https://your-app.com/api/talo/webhook",
		Motive:         "Order #12345",
		ClientData: &talo.ClientData{
			FirstName: "Juan",
			LastName:  "Perez",
			Email:     "juan@example.com",
			DNI:       "12345678",
		},
		// PartnerID: os.Getenv("TALO_PARTNER_ID"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Payment created:", payment.ID, payment.PaymentStatus)
}
```

## Authentication

The client automatically manages access tokens by calling `POST /users/{userId}/tokens` using your `clientId` + `clientSecret`.

- Tokens are cached in memory.
- Refreshed before expiration (using JWT `exp` claim).
- Automatically retries once on HTTP 401.

## Environments

```go
client, _ := talo.NewClient(talo.Config{
    // ...
    Environment: talo.Sandbox, // default is Production
    // BaseURL: "https://custom.example.com", // overrides environment
})
```

## Resources

### Payments

```go
payment, _ := client.Payments.Create(ctx, req)
payment, _ := client.Payments.Get(ctx, "PAY-xxx")
payment, _ := client.Payments.UpdateMetadata(ctx, "PAY-xxx", talo.UpdatePaymentMetadataRequest{Motive: "new reason"})
```

### Customers

```go
customer, _ := client.Customers.Create(ctx, talo.CreateCustomerRequest{...})
customer, _ := client.Customers.Get(ctx, "CUST-xxx")
tx, _ := client.Customers.GetTransaction(ctx, "CUST-xxx", "TX-xxx")
```

### Partners

```go
// Build onboarding redirect URL
authURL := client.Partners.GetAuthorizationURL("partner_123", "external_user_456")

// Exchange code from callback
exchange, _ := client.Partners.ExchangeToken(ctx, talo.PartnerTokenExchangeRequest{
    Code:        "code_from_redirect",
    ClientID:    os.Getenv("TALO_PARTNER_ID"),
    ClientSecret: os.Getenv("TALO_PARTNER_SECRET"),
})

// Account config
account, _ := client.Partners.GetAccount(ctx, exchange.UserID)
_, _ = client.Partners.UpdateAccount(ctx, exchange.UserID, talo.UpdatePartnerAccountRequest{
    TransferTolerance: intPtr(15),
})
```

### Refunds

```go
refund, _ := client.Refunds.Create(ctx, "PAY-123", talo.CreateRefundRequest{
    RefundType: "PARTIAL",
    Amount:     "500.00",
    Currency:   "ARS",
    Blame: talo.RefundBlame{
        TeamID: "support",
        Mail:   "support@your-app.com",
    },
    UserID: os.Getenv("TALO_USER_ID"),
})
```

### Sandbox

```go
// Only works in sandbox
resp, _ := client.Sandbox.SimulateCvuTransfer(ctx, "0000000000000000000000", talo.SimulateFaucetRequest{
    Amount: "1000.00",
})
```

## Webhooks (example)

```go
http.HandleFunc("/webhook", client.WebhookHandler(func(event *talo.WebhookEvent, payment *talo.Payment, r *http.Request) error {
    log.Printf("Payment %s updated to %s", event.PaymentID, event.Status)
    if payment != nil {
        log.Printf("Latest status from API: %s", payment.PaymentStatus)
    }
    return nil
}))
```

Or use the low-level parser:

```go
event, err := talo.ParseWebhook(rawBody)
```

## Error Handling

All API errors return `*talo.TaloError`:

```go
if taloErr, ok := err.(*talo.TaloError); ok {
    fmt.Println(taloErr.StatusCode, taloErr.Message, taloErr.RequestID)
}
```

## Design Notes

- Idiomatic Go: context.Context everywhere, pointer returns for responses, services for resources.
- Zero external dependencies (stdlib only).
- Automatic token management with JWT expiration parsing.
- 401 retry with token refresh (matches original TS behavior).
- Flexible JSON fields use `json.RawMessage` or `map[string]any` where the original used `.passthrough()`.

This SDK aims for feature parity with the official TypeScript SDK while following Go conventions.

For full API reference, see the original docs in https://github.com/talo-pay/talo-sdk
