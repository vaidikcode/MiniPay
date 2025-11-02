# MiniPay - Lightweight Payment Orchestration API

A production-grade Go REST API for managing payments with Stripe-inspired architecture. Implements charge creation, refunds, webhook delivery, idempotency, and transaction persistence using SQLite.

## Features

- **Charge Creation**: Atomic transaction creation with unique idempotency keys
- **Refunds**: Revert completed transactions with balance recalculation
- **Balance Tracking**: Real-time balance calculation with refund deductions
- **Webhook Delivery**: Async webhook processing with exponential backoff retries
- **Idempotency**: In-memory and DB-backed idempotency to prevent duplicate processing
- **Metrics**: Real-time metrics endpoint for monitoring charges, refunds, and webhook health
- **Concurrency Safe**: Thread-safe using sync.RWMutex and atomic operations
- **Health Checks**: Built-in health endpoint for liveness probes

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/vaidikcode/minipay.git
cd minipay
```

2. Download dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Create Charge

```bash
curl -X POST http://localhost:8080/api/v1/charges \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-unique-key-123" \
  -d '{
    "amount": 1000,
    "currency": "usd",
    "customer": "cust_123"
  }'
```

### Refund Transaction

```bash
curl -X POST http://localhost:8080/api/v1/refunds \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "txn_550e8400e29b41d4a716446655440000"
  }'
```

### Get Balance

```bash
curl http://localhost:8080/api/v1/balance
```

### Metrics

```bash
curl http://localhost:8080/metrics
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Testing

Run the full test suite:

```bash
go test -v ./...
```

Run with coverage:

```bash
go test -v -coverprofile=coverage.out ./...
```

## Docker

Build and run with Docker:

```bash
docker build -t minipay:latest .
docker run -p 8080:8080 minipay:latest
```
