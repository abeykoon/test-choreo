# Connections — Ordering & Payment

Two minimal Go HTTP services that demonstrate a service-to-service call. The
**ordering-service** receives an order request from a client and calls the
**payment-service** to process payment, attaching request context (headers +
`context.Context` timeout) on the outbound call.

```
client ──► ordering-service (:9090)  ──HTTP──►  payment-service (:9091)
   [Public]    POST /ordering/orders   [Project]   POST /payment/approved
```

---

## Endpoint visibility

Each service ships a `.choreo/component.yaml` (schema version `1.2`) that
declares the endpoint exposed by the service and its **network visibility**.

| Service            | Endpoint name | Base path   | Port  | Visibility    |
| ------------------ | ------------- | ----------- | ----- | ------------- |
| ordering-service   | `ordering-api`| `/ordering` | 9090  | **`Public`**  |
| payment-service    | `payment-api` | `/payment`  | 9091  | **`Project`** |

- **`Public`** — `ordering-service` is reachable from outside the Choreo
  project (e.g. by external clients calling `POST /ordering/orders`).
- **`Project`** — `payment-service` is only reachable from other components
  inside the *same* Choreo project. External callers cannot hit it directly;
  only `ordering-service` (in-project) can invoke `/payment/approved`.

This visibility split is what enforces the design: clients talk to the order
API, payment stays internal.

Accepted values for `networkVisibilities` in component.yaml: `Project`,
`Organization`, `Public` (default).

---

## Services

### 1. payment-service (port `9091`, context `/payment`, visibility **`Project`**)

A simple payment processor. Accepts a payment request and replies with an
approval or decline status. Exposed only inside the Choreo project — not
reachable from outside.

| Method | Path                  | Description                                       |
| ------ | --------------------- | ------------------------------------------------- |
| POST   | `/payment/approved`   | Marks the payment as `APPROVED` and returns a txn |
| POST   | `/payment/declined`   | Marks the payment as `DECLINED` and returns a txn |

**Headers read from the caller (context propagation):**

| Header              | Purpose                                |
| ------------------- | -------------------------------------- |
| `X-Correlation-ID`  | Echoed back in the response for tracing |
| `X-Source-Service`  | Logged so we know which service called |
| `X-Order-ID`        | Logged for correlation with orders     |

**Request body:**

```json
{ "orderId": "ord-1", "amount": 29.99 }
```

**Response body:**

```json
{
  "orderId": "ord-1",
  "status": "APPROVED",
  "transactionId": "txn-ord-1",
  "processedAt": "2026-05-13T12:00:00Z",
  "correlationId": "corr-ord-1-1715600000000000000"
}
```

---

### 2. ordering-service (port `9090`, context `/ordering`, visibility **`Public`**)

Exposes an endpoint to place an order. On each placed order it generates an
order ID + correlation ID, calls the payment service's `/payment/approved`
endpoint, and returns the combined result. Publicly reachable — this is the
service external clients hit.

| Method | Path                  | Description                                  |
| ------ | --------------------- | -------------------------------------------- |
| POST   | `/ordering/orders`    | Place an order; triggers a payment call      |

**Config (env var):**

| Variable               | Default                  | Purpose                          |
| ---------------------- | ------------------------ | -------------------------------- |
| `PAYMENT_SERVICE_URL`  | `http://localhost:9091`  | Base URL of the payment service  |

**Request body:**

```json
{ "item": "book", "quantity": 2, "amount": 29.99 }
```

**Response body:**

```json
{
  "order": {
    "id": "ord-1",
    "item": "book",
    "quantity": 2,
    "amount": 29.99,
    "status": "APPROVED"
  },
  "payment": {
    "orderId": "ord-1",
    "status": "APPROVED",
    "transactionId": "txn-ord-1",
    "processedAt": "2026-05-13T12:00:00Z",
    "correlationId": "corr-ord-1-1715600000000000000"
  }
}
```

---

## How ordering-service calls payment-service

The outbound HTTP call is built in `ordering-service/main.go` (`callPayment`).
It illustrates two ways context is attached to a downstream call:

1. **Go `context.Context` with timeout** — `http.NewRequestWithContext` is used
   with a 5-second timeout derived from the inbound request context. If the
   client disconnects or payment is slow, the call is cancelled.

   ```go
   ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
   defer cancel()
   req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
   ```

2. **HTTP headers carrying request metadata** — a correlation ID, source
   service name, and order ID are sent so the callee can trace the call back
   to its origin.

   ```go
   req.Header.Set("X-Correlation-ID", correlationID)
   req.Header.Set("X-Source-Service", "ordering-service")
   req.Header.Set("X-Order-ID", order.ID)
   ```

The target URL is `${PAYMENT_SERVICE_URL}/payment/approved`.

---

## Running locally

Open two terminals and run each service:

```bash
# terminal 1 — payment service
cd payment-service
go run .

# terminal 2 — ordering service
cd ordering-service
go run .
```

Optional — point the ordering service at a non-local payment service:

```bash
PAYMENT_SERVICE_URL=http://payment.internal:9091 go run .
```

---

## Invoking the services

**Place an order (the typical end-to-end flow):**

```bash
curl -X POST http://localhost:9090/ordering/orders \
  -H 'Content-Type: application/json' \
  -d '{"item":"book","quantity":2,"amount":29.99}'
```

**Call payment directly (bypasses ordering):**

```bash
curl -X POST http://localhost:9091/payment/approved \
  -H 'Content-Type: application/json' \
  -H 'X-Correlation-ID: test-123' \
  -H 'X-Source-Service: curl' \
  -d '{"orderId":"ord-99","amount":10.00}'
```

```bash
curl -X POST http://localhost:9091/payment/declined \
  -H 'Content-Type: application/json' \
  -d '{"orderId":"ord-99","amount":10.00}'
```

---

## Project layout

```
connections/
├── README.md
├── ordering-service/
│   ├── .choreo/
│   │   └── component.yaml   # endpoint: /ordering on :9090, visibility: Public
│   ├── go.mod
│   └── main.go              # exposes /ordering/orders; calls payment-service
└── payment-service/
    ├── .choreo/
    │   └── component.yaml   # endpoint: /payment on :9091, visibility: Project
    ├── go.mod
    └── main.go              # exposes /payment/approved and /payment/declined
```
