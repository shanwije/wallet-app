include .env

.PHONY: help up build down status logs clean migrate docs test test-unit test-integration fmt vet

# Help command for listing all available commands
help:
	@echo "Wallet App - Makefile Commands"
	@echo "----------------------------------------------------"
	@echo "  up         Start all services with fresh build"
	@echo "  build      Build containers without starting"
	@echo "  down       Stop all running services"
	@echo "  status     Show container status"
	@echo "  logs       Tail all logs from services"
	@echo "  clean      Stop and remove containers and volumes"
	@echo "  migrate    Run Goose DB migrations"
	@echo "  docs       Generate Swagger docs (requires swag)"
	@echo "  test       Run all tests (unit + integration)"
	@echo "  test-unit  Run unit tests only"
	@echo "  test-integration  Run integration tests only"
	@echo "  fmt        Format Go code"
	@echo "  vet        Run go vet for code analysis"
	@echo "----------------------------------------------------"

# Docker Compose - Start services (always rebuild)
up: fmt vet test-unit
	go mod tidy && docker compose -f deployments/docker-compose.yaml up -d --build --force-recreate

# Docker Compose - Build containers only
build: fmt vet test-unit
	go mod tidy && docker compose -f deployments/docker-compose.yaml build --no-cache

# Docker Compose - Stop services
down:
	docker compose -f deployments/docker-compose.yaml down

# Docker Compose - Show container status
status:
	docker compose -f deployments/docker-compose.yaml ps

# Docker Compose - View service logs
logs:
	docker compose -f deployments/docker-compose.yaml logs -f

# Docker Compose - Clean everything
clean:
	docker compose -f deployments/docker-compose.yaml down -v

# 🧪 Goose DB Migrations (ensure .env or ENV vars are available)
migrate:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir db/migrations postgres \
		"host=$$DB_HOST port=$$DB_PORT user=$$DB_USER password=$$DB_PASSWORD dbname=$$DB_NAME sslmode=$$DB_SSLMODE" up

# 📚 Swagger Docs (assumes swag installed globally)
docs:
	swag init -g cmd/main.go -o docs

# 🧪 Testing Commands
test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/... ./pkg/...

test-integration:
	@echo "Running integration tests..."
	@echo "Note: Integration tests require running services (make up first)"
	go test -v ./tests/integration/...

# 🔧 Code Quality Commands
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...
