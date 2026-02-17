# Employee Management Service

A simple REST API service for managing employee records, built with Go. Employee data is stored in memory.

## Project Structure

```
employee-mgt-service/
├── go.mod          # Go module file
├── main.go         # Service implementation
├── openapi.yaml    # OpenAPI 3.0 specification
└── README.md       # This file
```

## API Endpoints

All endpoints are prefixed with `/employee-mgt` context path.

| Method | Endpoint                        | Description           |
|--------|---------------------------------|-----------------------|
| GET    | `/employee-mgt/employees`       | Get all employees     |
| POST   | `/employee-mgt/employees`       | Add a new employee    |
| GET    | `/employee-mgt/employees/{id}`  | Get employee by ID    |

## Employee Schema

| Field      | Type   | Required | Description                    |
|------------|--------|----------|--------------------------------|
| id         | int    | No       | Auto-generated unique ID       |
| firstName  | string | Yes      | First name of the employee     |
| lastName   | string | Yes      | Last name of the employee      |
| email      | string | Yes      | Email address of the employee  |
| department | string | No       | Department name                |
| position   | string | No       | Job position or title          |

## Running the Service

```bash
cd employee-mgt-service
go run main.go
```

The service starts on port `8080`.

## API Usage Examples

### Add a New Employee

```bash
curl -X POST http://localhost:8080/employee-mgt/employees \
  -H "Content-Type: application/json" \
  -d '{
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "department": "Engineering",
    "position": "Software Engineer"
  }'
```

**Response:**
```json
{
  "id": 1,
  "firstName": "John",
  "lastName": "Doe",
  "email": "john.doe@example.com",
  "department": "Engineering",
  "position": "Software Engineer"
}
```

### Get All Employees

```bash
curl http://localhost:8080/employee-mgt/employees
```

**Response:**
```json
[
  {
    "id": 1,
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "department": "Engineering",
    "position": "Software Engineer"
  }
]
```

### Get Employee by ID

```bash
curl http://localhost:8080/employee-mgt/employees/1
```

**Response:**
```json
{
  "id": 1,
  "firstName": "John",
  "lastName": "Doe",
  "email": "john.doe@example.com",
  "department": "Engineering",
  "position": "Software Engineer"
}
```

### Error Responses

**Employee not found (404):**
```bash
curl http://localhost:8080/employee-mgt/employees/999
```

```json
{
  "message": "Employee not found",
  "code": 404
}
```

**Invalid request (400):**
```bash
curl -X POST http://localhost:8080/employee-mgt/employees \
  -H "Content-Type: application/json" \
  -d '{"firstName": "John"}'
```

```json
{
  "message": "FirstName, LastName, and Email are required",
  "code": 400
}
```

## OpenAPI Specification

The API is documented in `openapi.yaml` following the OpenAPI 3.0 specification. You can import this file into tools like Swagger UI or Postman for interactive API documentation.
