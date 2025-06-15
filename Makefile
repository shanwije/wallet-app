.PHONY: help up down status logs clean

help:
	@echo "Wallet App Commands:"
	@echo "  up      - Start all services"
	@echo "  down    - Stop all services"  
	@echo "  status  - Show container status"
	@echo "  logs    - Show all logs"
	@echo "  clean   - Stop and remove containers"
	@echo "  docs    - Generate Swagger documentation"

up: ## Start all services
	docker compose -f deployments/docker-compose.yaml up -d

down: ## Stop all services
	docker compose -f deployments/docker-compose.yaml down

status: ## Show container status
	docker compose -f deployments/docker-compose.yaml ps

logs: ## Show all logs
	docker compose -f deployments/docker-compose.yaml logs -f

clean: ## Stop and remove containers
	docker compose -f deployments/docker-compose.yaml down -v

docs: ## Generate Swagger documentation
	/Users/shan/go/bin/swag init -g cmd/main.go -o docs
