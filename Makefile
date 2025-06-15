include .env

.PHONY: help up build down status logs clean migrate docs

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
	@echo "----------------------------------------------------"

# Docker Compose - Start services (always rebuild)
up:
	go mod tidy && docker compose -f deployments/docker-compose.yaml up -d --build --force-recreate

# Docker Compose - Build containers only
build:
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
		"host=$$DB_HOST port=$$DB_PORT user=$$DB_USER password=$$DB_PASSWORD dbname=$$DB_NAME sslmode=disable" up

# 📚 Swagger Docs (assumes swag installed globally)
docs:
	swag init -g cmd/main.go -o docs
