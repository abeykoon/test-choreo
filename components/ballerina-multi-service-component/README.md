# Ballerina Multi-Service Component

A single Ballerina program with two services running on different ports.

## Services

| Service | Port | Context | Endpoint |
|---------|------|---------|----------|
| service | 9091 | `/wso2/services` | `GET /wso2/services/greeting` |
| library-service | 9092 | `/wso2/services/library` | `GET /wso2/services/library/books` |

## Run

```bash
bal run
```

Both services start together as a single program.

## Test

### Service

```bash
curl http://localhost:9091/wso2/services/greeting
```

Response:

```
Hello from wso2/services!
```

### Library Service

```bash
curl http://localhost:9092/wso2/services/library/books
```

Response:

```
Hello from wso2/services/library!
```
