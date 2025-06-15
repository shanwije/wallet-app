# Wallet Backend Service - Crypto.com Assignment

A robust, production-ready centralized wallet backend service built with Go, following clean architecture principles and financial industry best practices. This project was developed as a coding assignment to demonstrate senior-level engineering capabilities.

## Requirements

### User Stories Implemented:
- **User can deposit money into wallet** - Complete with validation and ACID transactions
- **User can withdraw money from wallet** - Includes balance validation and atomic operations
- **User can send money to another user** - Atomic transfers between wallets
- **User can check wallet balance** - Real-time balance inquiry
- **User can view transaction history** - Complete audit trail with transaction types

### Technical Requirements
- **Language**: Go 1.21+
- **Database**: PostgreSQL
- **In-memory Database**: Redis-ready architecture (in-memory cache for development, later can easily configured through current cache interface)
- **Centralized Wallet**: Complete user and wallet management system

## Design Decisions

#### 1. **Financial Precision**
- **Decision**: Use `github.com/shopspring/decimal` for all monetary calculations
- **Implementation**: All money values stored as `DECIMAL(20,2)` in database

#### 2. **ACID Transaction Compliance**
- **Decision**: Wrap all financial operations in database transactions
- **Implementation**: Service layer manages transaction boundaries with proper rollback

#### 3. **Double-Entry Transaction Recording**
- **Decision**: Create transaction records for both sides of transfers
- **Implementation**: Linked transactions using `reference_id` field

#### 6. **Idempotency Support**
- **Decision**: Implement idempotency middleware for POST operations
- **Implementation**: In-memory cache for development, Redis or Memcache ready architecture for production

## Quick Start Guide

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Make (optional)

### Option 1: Docker Compose (Recommended)

1. **Clone and setup**
   ```bash
   git clone <repository-url>
   cd wallet-app
   cp .env.example .env
   ```

2. **Start all services**
   ```bash
   make up
   # This will:
   # - Build the application
   # - Start PostgreSQL database
   # - Run database migrations
   # - Start the API server
   ```

3. **Verify the setup**
   ```bash
   curl http://localhost:8082/health
   # Expected: {"status":"ok","database":"connected"}
   ```

4. **Access the API**
   - **API Base URL**: http://localhost:8082/api/v1
   - **Swagger Documentation**: http://localhost:8082/swagger/index.html
   - **Health Check**: http://localhost:8082/health

### Option 2: Local Configuration

1. **Setup PostgreSQL**
   ```bash
   # Using Docker
   docker run -d \
     --name wallet-postgres \
     -e POSTGRES_USER=wallet \
     -e POSTGRES_PASSWORD=walletpass \
     -e POSTGRES_DB=wallet_db \
     -p 5432:5432 \
     postgres:15
   ```

2. **Run migrations**
   ```bash
   make migrate
   ```

3. **Start the application**
   ```bash
   make run
   # or directly: go run cmd/main.go
   ```

##  Testing Strategy

### Test Coverage Overview
- **Unit Tests**: Service layer business logic (60%+ coverage, This could have been even higher if the scope of the repository is larger )
- **Integration Tests**: Full API workflow testing

### Running Tests

```bash
# Run all tests
make test

# Unit tests only
make test-unit

# Integration tests only  
make test-integration

# With coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## API Endpoints ( Please refer to swagger for more information )

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



#### **Additional Features** (Beyond requirements)

- Swagger API documentation
- Idempotency middleware
- Structured logging with request tracing
- Health check endpoints
- Docker containerization
- Database migrations
- Comprehensive error handling
- Configuration validation

#### **Features Not Implemented** (Conscious decisions)
- **Authentication/Authorization**: Not required within the scope, noted for production
- **Redis/ Memcache Integration**: Architecture ready with Docker container, using in-memory cache for simplicity  
- **Rate Limiting**: Production feature, not core to wallet functionality
- **Pagination**: Transaction history returns all records (easily extendable)
- **Audit Logging**: Basic transaction records implemented, advanced auditing for production

## 🛠️ Areas for Improvement

### **Short-term Enhancements** (Production readiness)
1. **Authentication & Authorization**
   - JWT-based user authentication
   - Role-based access control (RBAC)
   - API key management for service-to-service communication

2. **Enhanced Security**
   - Rate limiting per user/IP
   - Request signing for sensitive operations
   - Audit logging for compliance
   - Data encryption at rest

3. **Performance Optimization**
   - Transaction history pagination
   - Database query optimization
   - Redis implementation for distributed idempotency
   - Connection pool tuning

4. **Ops Support**
   - Metrics collection (Prometheus)
   - Distributed tracing (Jaeger)
   - Alerting and monitoring dashboards
   - Log aggregation (ELK stack)

### **Long-term Enhancements** (Scalability)
1. **Microservices Architecture**
   - Separate user service and wallet service
   - Event-driven architecture with message queues
   - API Gateway for unified interface

2. **Data Layer**
   - Read replicas for query performance
   - Database sharding for horizontal scaling
   - Event sourcing for complete audit trail

3. **Advanced Features**
   - Multi-currency support
   - Transaction categorization and tagging
   - Spending limits and controls
   - Scheduled/recurring transactions

## 🔧 Technical Implementation Details

### **Project Structure**
```
wallet-app/
├── cmd/main.go                 # Application bootstrap
├── internal/                   # Private application code
│   ├── api/                    # HTTP layer
│   │   ├── handlers/           # Request handlers
│   │   └── router.go           # Route configuration
│   ├── config/                 # Configuration management
│   ├── middleware/             # HTTP middleware
│   ├── models/                 # Domain models
│   ├── repository/             # Data access layer
│   │   └── postgres/           # PostgreSQL implementations
│   └── service/                # Business logic layer
├── pkg/                        # Reusable packages
│   ├── db/                     # Database utilities
│   ├── errors/                 # Error handling
│   ├── health/                 # Health checks
│   └── logger/                 # Logging utilities
├── tests/integration/          # Integration tests
├── db/migrations/              # Database schema
├── deployments/                # Docker configuration
└── docs/                       # API documentation
```

## 📊 API Examples

### **Create User with Wallet**
```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: user-creation-001" \
  -d '{"name": "John Doe"}'

# Response:
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "created_at": "2024-06-16T10:30:00Z"
  },
  "wallet": {
    "id": "456e7890-e89b-12d3-a456-426614174001",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "balance": "0.00",
    "created_at": "2024-06-16T10:30:00Z"
  }
}
```

### **Deposit Funds**
```bash
curl -X POST http://localhost:8082/api/v1/wallets/456e7890-e89b-12d3-a456-426614174001/deposit \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: deposit-001" \
  -d '{"amount": 100.50}'

# Response:
{
  "id": "456e7890-e89b-12d3-a456-426614174001",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "balance": "100.50",
  "created_at": "2024-06-16T10:30:00Z"
}
```

### **Transfer Between Wallets**
```bash
curl -X POST http://localhost:8082/api/v1/wallets/456e7890-e89b-12d3-a456-426614174001/transfer \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: transfer-001" \
  -d '{
    "to_wallet_id": "789e0123-e89b-12d3-a456-426614174002",
    "amount": 25.00,
    "description": "Payment for services"
  }'

# Response: HTTP 200 OK (no body for transfer operations)
```

### **Get Transaction History**
```bash
curl http://localhost:8082/api/v1/wallets/456e7890-e89b-12d3-a456-426614174001/transactions

# Response:
[
  {
    "id": "tx-001",
    "wallet_id": "456e7890-e89b-12d3-a456-426614174001",
    "type": "deposit",
    "amount": "100.50",
    "description": null,
    "created_at": "2024-06-16T10:30:00Z"
  },
  {
    "id": "tx-002",
    "wallet_id": "456e7890-e89b-12d3-a456-426614174001",
    "type": "transfer_out",
    "amount": "25.00",
    "reference_id": "ref-001",
    "description": "Payment for services",
    "created_at": "2024-06-16T10:35:00Z"
  }
]
```

## 🏃‍♂️ Makefile Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `make help` | Show all available commands | Development guidance |
| `make up` | Start all services (build + migrate + run) | Full environment setup |
| `make down` | Stop all services | Environment cleanup |
| `make build` | Build containers without starting | CI/CD builds |
| `make status` | Show container status | Environment monitoring |
| `make logs` | View service logs | Debugging |
| `make clean` | Stop and remove all containers + volumes | Full cleanup |
| `make migrate` | Run database migrations | Schema updates |
| `make test` | Run all tests (unit + integration) | Quality assurance |
| `make test-unit` | Run unit tests only | Fast feedback loop |
| `make test-integration` | Run integration tests only | API validation |
| `make fmt` | Format Go code | Code consistency |
| `make vet` | Run go vet analysis | Static analysis |
| `make docs` | Generate Swagger documentation | API docs |

## 📝 API Documentation

### **OpenAPI/Swagger**
- **Interactive Docs**: Available at `/swagger/index.html` when running
- **Specification**: Generated from Go code annotations
- **Testing Interface**: Direct API testing from documentation

### **Endpoint Summary**

| Method | Endpoint | Purpose | Request Body | Response |
|--------|----------|---------|--------------|----------|
| POST | `/api/v1/users` | Create user + wallet | `{"name": "string"}` | User + Wallet objects |
| POST | `/api/v1/wallets/{id}/deposit` | Add funds | `{"amount": number}` | Updated wallet |
| POST | `/api/v1/wallets/{id}/withdraw` | Remove funds | `{"amount": number}` | Updated wallet |
| POST | `/api/v1/wallets/{id}/transfer` | Send to another wallet | `{"to_wallet_id": "uuid", "amount": number, "description": "string"}` | Success status |
| GET | `/api/v1/wallets/{id}/balance` | Check balance | None | Wallet object |
| GET | `/api/v1/wallets/{id}/transactions` | Transaction history | None | Transaction array |
| GET | `/health` | Service health | None | Health status |

### **Error Response Format**
```json
{
  "error": "descriptive error message",
  "code": 400
}
```

### **Idempotency Header**
```bash
# All POST requests support idempotency
Idempotency-Key: unique-operation-identifier
```