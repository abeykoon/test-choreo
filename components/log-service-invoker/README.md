# Log Service Invoker

A Go service that invokes the downstream `log-service` `POST /test/greeting`
endpoint with a caller-supplied JSON payload. The invocation goes through the
Choreo API Manager using OAuth2 client credentials and an optional
Choreo-API-Key.

## Project Structure

```
log-service-invoker/
├── .choreo/
│   └── component.yaml   # Choreo component configuration
├── go.mod               # Go module file
├── go.sum               # Go module checksums
├── main.go              # Service implementation
├── openapi.yaml         # OpenAPI 3.0 specification
└── README.md            # This file
```

## API Endpoints

| Method | Endpoint    | Description                                    |
|--------|-------------|------------------------------------------------|
| POST   | `/invoker`  | Forwards the JSON payload to the log service   |

## Configuration

The downstream log-service connection is injected via the following
environment variables (populated by the Choreo connection reference):

| Variable                                        | Description                          |
|-------------------------------------------------|--------------------------------------|
| `CHOREO_LOG_SVC_CONNECTION_CONSUMERKEY`     | OAuth2 client ID                     |
| `CHOREO_LOG_SVC_CONNECTION_CONSUMERSECRET`  | OAuth2 client secret                 |
| `CHOREO_LOG_SVC_CONNECTION_SERVICEURL`      | Base URL of the log service          |
| `CHOREO_LOG_SVC_CONNECTION_TOKENURL`        | OAuth2 token endpoint                |
| `CHOREO_LOG_SVC_CONNECTION_APIKEY`          | Choreo-API-Key (optional)            |

## Running the Service

```bash
cd log-service-invoker
go run main.go
```

The service starts on port `9090`.

## API Usage

### POST `/invoker`

```bash
curl -i -X POST http://localhost:9090/invoker \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello",
    "from": "invoker"
  }'
```

**Response:**

```json
{
  "status": 200,
  "correlationId": "corr-invoker-1720000000000000000",
  "upstreamStatus": 200,
  "payload": {
    "message": "Hello",
    "from": "invoker"
  },
  "response": {
    "message": "Hello",
    "from": "invoker"
  }
}
```
