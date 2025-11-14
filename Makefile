.PHONY: help build run test migrate-up migrate-down migrate-create migrate-status clean docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make migrate-up     - Run database migrations up"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make migrate-create - Create a new migration file"
	@echo "  make migrate-status - Show migration status"
	@echo "  make docker-up      - Start Docker containers"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make clean          - Clean build artifacts"

build:
	@go build -o bin/server ./cmd/server

run:
	@go run ./cmd/server

test:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

migrate-up:
	@goose -dir migrations postgres "host=localhost port=5432 user=flypro_user password=flypro_password dbname=flypro_db sslmode=disable" up

migrate-down:
	@goose -dir migrations postgres "host=localhost port=5432 user=flypro_user password=flypro_password dbname=flypro_db sslmode=disable" down

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir migrations postgres "host=localhost port=5432 user=flypro_user password=flypro_password dbname=flypro_db sslmode=disable" create $$name sql

migrate-status:
	@goose -dir migrations postgres "host=localhost port=5432 user=flypro_user password=flypro_password dbname=flypro_db sslmode=disable" status

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down

clean:
	@rm -rf bin/
	@rm -f coverage.out coverage.html

