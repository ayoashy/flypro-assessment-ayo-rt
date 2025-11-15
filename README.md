
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


### 1. How would you handle concurrent expense approvals?

**Strategy**: Implement optimistic locking using version fields and database transactions.

```go
// Add version field to Expense model
type Expense struct {
    // ... existing fields
    Version int `gorm:"default:0"`
}

// In service layer
func (s *ExpenseService) ApproveExpense(ctx context.Context, id uint, version int) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        var expense Expense
        if err := tx.Where("id = ? AND version = ?", id, version).
            First(&expense).Error; err != nil {
            return errors.New("expense was modified")
        }

        expense.Status = "approved"
        expense.Version++
        return tx.Save(&expense).Error
    })
}
```

**Alternative**: Use database-level locking:

```sql
SELECT * FROM expenses WHERE id = ? FOR UPDATE;
```

### 2. What strategies would you use to scale this system?

**Horizontal Scaling**:

- Stateless API servers (multiple instances behind load balancer)
- Read replicas for PostgreSQL
- Redis cluster for distributed caching
- Sharding by user_id for very large datasets

**Vertical Scaling**:

- Connection pooling (already implemented)
- Database query optimization with indexes
- Redis caching to reduce database load

**Microservices Approach**:

- Separate services for: User Management, Expense Management, Reporting, Currency Conversion
- Event-driven architecture with message queue (Kafka/RabbitMQ)
- API Gateway for routing

**Caching Strategy**:

- Multi-layer caching (L1: in-memory, L2: Redis, L3: Database)
- Cache warming for frequently accessed data
- Cache invalidation strategies

**Background Jobs**:

- Async processing for currency conversion
- Background report generation
- Email notifications

### 3. How would you ensure data consistency across services?

**ACID Transactions**:

- Use database transactions for critical operations
- Implement compensating transactions for distributed scenarios

**Event Sourcing**:

- Store all changes as events
- Rebuild state from events
- Enable audit trail and replay

**Saga Pattern**:

- For distributed transactions across services
- Compensating actions for rollback

**Idempotency**:

- Idempotency keys for all mutations
- Idempotent API endpoints

**Eventual Consistency**:

- Accept eventual consistency for non-critical paths
- Use eventual consistency with conflict resolution for reporting

### 4. What monitoring and alerting would you implement?

**Metrics**:

- Prometheus metrics for:
  - Request rate (requests/sec)
  - Error rate (4xx, 5xx errors)
  - Response time (p50, p95, p99)
  - Database connection pool usage
  - Redis cache hit/miss ratio
  - Currency API call success rate

**Logging**:

- Structured logging with correlation IDs
- Log levels: DEBUG, INFO, WARN, ERROR
- Centralized logging (ELK Stack or similar)

**Health Checks**:

- `/health` endpoint
- Database connectivity check
- Redis connectivity check
- External API (currency) health check

**Alerting**:
- Database connection pool exhaustion
- Redis unavailability
- Currency API failures

**Tracing**:
- Performance bottleneck identification

