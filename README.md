# Travel Expense Management API

A comprehensive backend API for managing travel expenses, built with Go, Gin, GORM, PostgreSQL, and Redis.

## Features

- User management with email validation
- Expense tracking with multi-currency support
- Expense report creation and submission
- Currency conversion with Redis caching
- Comprehensive input validation
- Database migrations with Goose
- RESTful API design
- Structured logging
- Error handling middleware

## Tech Stack

- **Language**: Go 1.25
- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Migrations**: Goose
- **Validation**: go-playground/validator

## Prerequisites

- Go 1.25 or higher
- Docker and Docker Compose
- Make (optional, for convenience commands)

## Quick Start

### 1. Start Services

```bash
make docker-up
```

This starts PostgreSQL and Redis containers.

### 2. Run Migrations

```bash
make migrate-up
```

### 3. Set Environment Variables

Create a `.env` file in the root directory:

```env
SERVER_PORT=8080
SERVER_HOST=localhost
DB_HOST=localhost
DB_PORT=5432
DB_USER=flypro_user
DB_PASSWORD=flypro_password
DB_NAME=flypro_db
DB_SSLMODE=disable
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
CURRENCY_API_URL=https://api.exchangerate-api.com/v4/latest
ENV=development
```

### 4. Run the Application

```bash
make run
```

Or:

```bash
go run ./cmd/server
```

The API will be available at `http://localhost:8080`

## API Endpoints

### User Management

- `POST /api/users` - Create a new user
- `GET /api/users/:id` - Get user details

### Expense Management

- `POST /api/expenses` - Create an expense
- `GET /api/expenses` - List expenses (with pagination and filtering)
- `GET /api/expenses/:id` - Get expense details
- `PUT /api/expenses/:id` - Update an expense
- `DELETE /api/expenses/:id` - Delete an expense

### Expense Reports

- `POST /api/reports` - Create an expense report
- `GET /api/reports` - List reports (with pagination)
- `GET /api/reports/:id` - Get report details
- `POST /api/reports/:id/expenses` - Add expenses to a report
- `PUT /api/reports/:id/submit` - Submit a report for approval

## Request Headers

For expense and report operations, include the user ID in the header:

```
X-User-ID: 1
```

## Example Requests

### Create User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "name": "John Doe"
  }'
```

### Create Expense

```bash
curl -X POST http://localhost:8080/api/expenses \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "amount": 150.50,
    "currency": "USD",
    "category": "travel",
    "description": "Flight ticket"
  }'
```

### Create Report

```bash
curl -X POST http://localhost:8080/api/reports \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "title": "Q1 2024 Travel Expenses"
  }'
```

## Database Migrations

### Run Migrations

```bash
make migrate-up
```

### Rollback Last Migration

```bash
make migrate-down
```

### Check Migration Status

```bash
make migrate-status
```

### Create New Migration

```bash
make migrate-create
```

## Testing

Run tests with coverage:

```bash
make test
```

This generates `coverage.out` and `coverage.html` files.

## Project Structure

```
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/          # Configuration management
│   ├── dto/             # Data transfer objects
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Database models
│   ├── repository/      # Data access layer
│   ├── services/        # Business logic
│   ├── utils/           # Utility functions
│   └── validators/      # Custom validators
├── migrations/          # Database migrations
├── tests/               # Test files
├── docker-compose.yml   # Docker services
├── Makefile            # Build commands
└── README.md           # This file
```

## Architecture

The application follows a layered architecture:

1. **Handler Layer**: HTTP request/response handling
2. **Service Layer**: Business logic and orchestration
3. **Repository Layer**: Data access abstraction
4. **Model Layer**: Domain entities


## Currency Conversion

The API integrates with ExchangeRate-API for currency conversion. Exchange rates are cached in Redis with a 6-hour TTL to reduce API calls and improve performance.

## Error Handling

The API returns consistent error responses:

```json
{
  "success": false,
  "error": {
    "code": 400,
    "message": "validation failed",
    "fields": {
      "email": "email must be a valid email"
    }
  }
}
```


## Scaling Considerations

### Concurrent Expense Approvals

To handle concurrent expense approvals, consider:

1. **Database Transactions**: optimistic locking with version fields
2. **Distributed Locks**: Redis for distributed locking
3. **Event Sourcing**: all state changes for audit and conflict resolution
4. **Queue System**: Process approvals asynchronously with message queues

### System Scaling

1. **Database**: 
   - Read replicas for read-heavy operations
   - Connection pooling
   - Query optimization with proper indexes

2. **Caching**:
   - Redis cluster for high availability
   - Cache warming strategies
   - Cache invalidation patterns

3. **Application**:
   - Horizontal scaling with load balancers
   - Stateless API design
   - Graceful shutdown 

4. **Background Jobs**:
   - Use message queues (RabbitMQ, Kafka) for async processing
   - Worker pools for currency conversion
   - Scheduled jobs for report generation

### Data Consistency

1. **Transactions**: Use database transactions for multi-step operations
2. **Saga Pattern**: For distributed transactions across services
3. **Eventual Consistency**: Accept eventual consistency for non-critical paths
4. **Idempotency**: Make operations idempotent with unique keys

### Monitoring and Alerting

1. **Metrics**:
   - Request rate and latency
   - Error rates by endpoint
   - Database connection pool usage
   - Cache hit/miss ratios
   - Currency API response times

3. **Alerting**:
   - High error rates
   - Slow response times (>1s)
   - Database connection failures
   - Redis unavailability
   - Currency API failures

4. **Health Checks**:
   - Database connectivity
   - Redis connectivity
   - External API health

