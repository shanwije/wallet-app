# Wallet Backend Service

A robust, production-ready centralized wallet backend service built with Go, following clean architecture principles and financial industry best practices.

## Features

### Core Functionality
- **User Management**: Create users with automatic wallet creation
- **Wallet Operations**: 
  - Deposit funds to wallet
  - Withdraw funds from wallet
  - Transfer funds between wallets
  - Check wallet balance
  - Get transaction history
- **Financial Compliance**: 
  - ACID-compliant transactions
  - Precise decimal arithmetic (no floating-point errors)
  - Double-entry transaction records
  - Atomic transfers with rollback capability

### Technical Features
- **Clean Architecture**: Separation of concerns with domain, service, and repository layers
- **Database Support**: PostgreSQL with migration support
- **API Documentation**: Swagger/OpenAPI 3.0 specification
- **Idempotency**: Prevent duplicate transactions with idempotency keys
- **Error Handling**: Comprehensive error handling and validation
- **Testing**: Unit tests and integration tests
- **Docker Support**: Containerized deployment
- **Health Checks**: Service health monitoring

## Architecture

```
cmd/                    # Application entry points
internal/
├── api/               # HTTP handlers and routing
│   └── handlers/      # Request handlers
├── config/            # Configuration management
├── domain/            # Business domain models
├── middleware/        # HTTP middleware (idempotency, etc.)
├── models/            # Data models
├── repository/        # Data access layer
│   └── postgres/      # PostgreSQL implementations
├── service/           # Business logic layer
└── utils/             # Utility functions
pkg/
└── db/                # Database connection and utilities
tests/
└── integration/       # Integration tests
```

## API Endpoints

### User Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/users` | Create new user with wallet |

### Wallet Operations
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/wallets/{id}/deposit` | Deposit funds |
| POST | `/api/v1/wallets/{id}/withdraw` | Withdraw funds |
| POST | `/api/v1/wallets/{id}/transfer` | Transfer to another wallet |
| GET | `/api/v1/wallets/{id}/balance` | Get wallet balance |
| GET | `/api/v1/wallets/{id}/transactions` | Get transaction history |

### System
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health check |
| GET | `/swagger/index.html` | API documentation |

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 12+
- Docker (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd wallet-app
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your database configuration
   ```

3. **Start PostgreSQL** (using Docker)
   ```bash
   make db-up
   ```

4. **Run database migrations**
   ```bash
   make migrate-up
   ```

5. **Run the application**
   ```bash
   make run
   ```

6. **Access the API**
   - API: http://localhost:8082/api/v1
   - Swagger: http://localhost:8082/swagger/index.html
   - Health: http://localhost:8082/health

### Using Docker

```bash
# Build and start all services
make up

# Stop services
make down

# View logs
make logs
```

## Testing

### Unit Tests
```bash
make test-unit
```

### Integration Tests
```bash
# Start the application first
make up

# Run integration tests
make test-integration
```

### All Tests
```bash
make test
```

## Money Handling

This service uses precise decimal arithmetic to handle money correctly:

- **No floating-point arithmetic**: Uses `github.com/shopspring/decimal` for all monetary calculations
- **Database storage**: Money stored as `DECIMAL(19,4)` for precision
- **API format**: Accepts money as `float64` in JSON but immediately converts to `decimal.Decimal`

## Transaction Guarantees

### ACID Compliance
- **Atomicity**: All wallet operations are wrapped in database transactions
- **Consistency**: Balance constraints and validation rules are enforced
- **Isolation**: Concurrent transactions use proper locking
- **Durability**: All operations are persisted to PostgreSQL

### Double-Entry Bookkeeping
- Every transfer creates two transaction records (debit and credit)
- Transaction records are linked via `reference_id`
- Complete audit trail for all monetary movements

## Idempotency

The service supports idempotency for all POST operations:

```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-123" \
  -d '{"name": "John Doe"}'
```

- Same `Idempotency-Key` returns identical response
- Prevents duplicate transactions
- Keys expire after 24 hours

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_PORT` | HTTP server port | `8082` |
| `API_VERSION` | API version prefix | `v1` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | Database name | `wallet_db` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `password` |
| `DB_SSLMODE` | SSL mode | `disable` |

## Error Handling

The API returns consistent error responses:

```json
{
  "error": "insufficient balance",
  "code": 400
}
```

Common HTTP status codes:
- `200`: Success
- `201`: Created
- `400`: Bad Request (validation error)
- `404`: Not Found
- `500`: Internal Server Error

## Security Considerations

### Current Implementation
- Input validation and sanitization
- SQL injection prevention using parameterized queries
- CORS headers configuration
- Request size limiting

### Production Recommendations
- Add authentication/authorization (JWT tokens)
- Rate limiting
- API key management
- Audit logging
- Encrypt sensitive data at rest
- Use HTTPS only
- Network segmentation

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Wallets Table
```sql
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    balance DECIMAL(19,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Transactions Table
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    type VARCHAR(50) NOT NULL,
    amount DECIMAL(19,4) NOT NULL,
    reference_id UUID,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Performance Considerations

- Database indexes on frequently queried columns
- Connection pooling for database connections
- Prepared statements for repeated queries
- Pagination support for transaction history (future enhancement)

## Monitoring and Observability

### Health Checks
- `/health` endpoint for service health
- Database connectivity check
- Response time monitoring

### Logging
- Structured logging with request IDs
- Error logging with stack traces
- Audit logs for financial operations

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make up` | Start all services with Docker |
| `make down` | Stop all services |
| `make build` | Build the application |
| `make run` | Run the application locally |
| `make test` | Run all tests |
| `make test-unit` | Run unit tests only |
| `make test-integration` | Run integration tests only |
| `make db-up` | Start PostgreSQL container |
| `make db-down` | Stop PostgreSQL container |
| `make migrate-up` | Apply database migrations |
| `make migrate-down` | Rollback database migrations |
| `make fmt` | Format Go code |
| `make vet` | Run go vet |
| `make docs` | Generate Swagger documentation |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make test` and `make fmt`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
