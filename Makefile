include .env

run:
	go run cmd/main.go

dev-up:
	docker compose -f deployments/docker-compose.yaml up --build

migrate:
	goose -dir db/migrations postgres "user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) host=$(DB_HOST) sslmode=disable" up
