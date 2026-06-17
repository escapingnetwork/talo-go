# Talo Pay SDK for Go

Official Go SDK for the Talo Transfers API, structured similarly to the [Mercado Pago Go SDK](https://github.com/mercadopago/sdk-go).

## Installation

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

	"github.com/escapingnetwork/talo-go/pkg/config"
	"github.com/escapingnetwork/talo-go/pkg/payment"
)

func main() {
	cfg, err := config.New(
		os.Getenv("TALO_CLIENT_ID"),
		os.Getenv("TALO_CLIENT_SECRET"),
		os.Getenv("TALO_USER_ID"),
		config.WithEnvironment(config.Sandbox),
	)
	if err != nil {
		log.Fatal(err)
	}

	paymentClient := payment.NewClient(cfg)

	pay, err := paymentClient.Create(context.Background(), payment.CreateRequest{
		UserID: os.Getenv("TALO_USER_ID"),
		Price: payment.CreatePaymentPrice{
			Amount:   1500,
			Currency: "ARS",
		},
		PaymentOptions: []string{"transfer"},
		ExternalID:     "order_12345",
		WebhookURL:     "https://your-app.com/webhook",
		Motive:         "Order #12345",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created payment:", pay.ID, pay.PaymentStatus)
}
```

## Configuration

```go
cfg, _ := config.New(clientID, clientSecret, userID,
    config.WithEnvironment(config.Sandbox),
    // config.WithBaseURL("https://custom.api.talo.com.ar"),
)
```

The `Config` automatically handles:
- JWT token acquisition & caching
- Token refresh before expiration
- Automatic retry on 401 Unauthorized

## Resources

Each resource lives in its own package under `pkg/` and follows the same pattern:

```go
import (
    "github.com/escapingnetwork/talo-go/pkg/config"
    "github.com/escapingnetwork/talo-go/pkg/payment"
    "github.com/escapingnetwork/talo-go/pkg/customer"
    "github.com/escapingnetwork/talo-go/pkg/partner"
    "github.com/escapingnetwork/talo-go/pkg/refund"
    "github.com/escapingnetwork/talo-go/pkg/sandbox"
)

cfg, _ := config.New(...)

paymentClient := payment.NewClient(cfg)
customerClient := customer.NewClient(cfg)
partnerClient := partner.NewClient(cfg)
refundClient := refund.NewClient(cfg)
sandboxClient := sandbox.NewClient(cfg)
```

### Payments

```go
p, _ := paymentClient.Create(ctx, payment.CreateRequest{...})
p, _ := paymentClient.Get(ctx, "PAY-xxx")
p, _ := paymentClient.UpdateMetadata(ctx, "PAY-xxx", payment.UpdateMetadataRequest{Motive: "new reason"})
```

### Customers

```go
c, _ := customerClient.Create(ctx, customer.CreateRequest{...})
c, _ := customerClient.Get(ctx, "CUST-xxx")
tx, _ := customerClient.GetTransaction(ctx, "CUST-xxx", "TX-xxx")
```

### Partners

```go
authURL := partnerClient.GetAuthorizationURL("partner_123", "ext_user_456")
exchange, _ := partnerClient.ExchangeToken(ctx, partner.TokenExchangeRequest{...})
account, _ := partnerClient.GetAccount(ctx, exchange.UserID)
_, _ = partnerClient.UpdateAccount(ctx, userID, partner.UpdateAccountRequest{...})
```

### Refunds

```go
r, _ := refundClient.Create(ctx, "PAY-123", refund.CreateRequest{
    RefundType: "PARTIAL",
    Amount:     "500.00",
    Currency:   "ARS",
    Blame:      refund.Blame{TeamID: "support", Mail: "support@your-app.com"},
    UserID:     os.Getenv("TALO_USER_ID"),
})
```

### Sandbox

```go
resp, _ := sandboxClient.SimulateCvuTransfer(ctx, cvu, sandbox.SimulateFaucetRequest{Amount: "1000"})
```

## Error Handling

```go
if taloErr, ok := err.(*config.TaloError); ok {
    fmt.Printf("Talo error: %s (status=%d, request_id=%s)\n", 
        taloErr.Message, taloErr.StatusCode, taloErr.RequestID)
}
```

## Project Structure

```
pkg/
├── config/           # Central configuration + token management + HTTP layer
├── payment/          # client.go + request.go + response.go
├── customer/
├── partner/
├── refund/
└── sandbox/
```

This structure closely follows the official Mercado Pago Go SDK pattern:
- One `Config` per set of credentials
- Independent `NewXXXClient(cfg)` for each domain
- Clear separation of request/response models
