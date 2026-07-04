# Log Service

A simple Go service that logs incoming request headers and payloads. It exposes GET and POST resources under the `/test` context path.

## Project Structure

```
log-service/
├── .choreo/
│   └── component.yaml   # Choreo component configuration
├── go.mod               # Go module file
├── main.go              # Service implementation
├── openapi.yaml         # OpenAPI 3.0 specification
└── README.md            # This file
```

## API Endpoints

All endpoints are prefixed with the `/test` context path.

| Method | Endpoint          | Description                                        |
|--------|-------------------|----------------------------------------------------|
| GET    | `/test/greeting`  | Returns a static `Hello World` string              |
| POST   | `/test/greeting`  | Echoes back the incoming JSON payload as JSON      |

Both endpoints log all incoming request headers to standard output. The POST endpoint additionally logs the incoming payload.

## Running the Service

```bash
cd log-service
go run main.go
```

The service starts on port `9090`.

## API Usage Examples

### GET `/test/greeting`

```bash
curl -i http://localhost:9090/test/greeting
```

**Response:**

```
HTTP/1.1 200 OK
Content-Type: text/plain

Hello World
```

### POST `/test/greeting`

```bash
curl -i -X POST http://localhost:9090/test/greeting \
  -H "Content-Type: application/json" \
  -H "X-Custom-Header: my-value" \
  -d '{
    "message": "Hello",
    "from": "client"
  }'
```

**Response:**

```json
{
  "message": "Hello",
  "from": "client"
}
```

### Error: Invalid JSON payload

```bash
curl -i -X POST http://localhost:9090/test/greeting \
  -H "Content-Type: application/json" \
  -d 'not-a-json'
```

**Response:**

```json
{
  "message": "Invalid JSON payload",
  "code": 400
}
```

## OpenAPI Specification

The API is documented in `openapi.yaml` following the OpenAPI 3.0 specification. Import this file into tools like Swagger UI or Postman for interactive API documentation.
