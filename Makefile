.PHONY: help docker-up docker-down migrate-up migrate-down test lint build run clean

include .env
export

# Variables
MYSQL_HOST ?= localhost
MYSQL_PORT ?= 3306
MYSQL_USER ?= slack_bot
MYSQL_PASSWORD ?= slack_bot_password
MYSQL_DATABASE ?= translation_bot
DATABASE_URL = mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

help:
	@echo "Available commands:"
	@echo "  make docker-up      - Start Docker containers (MySQL, Redis)"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linter (golangci-lint)"
	@echo "  make build          - Build binary"
	@echo "  make run            - Run the application"
	@echo "  make clean          - Clean build artifacts"

docker-up:
	docker-compose --env-file .env up -d
	@echo "Docker containers started. Waiting for MySQL to be ready..."
	@sleep 15

docker-down:
	docker-compose down

migrate-up: docker-up
	@echo "Running database migrations..."
	@which migrate > /dev/null || (echo "migrate CLI not found. Installing..."; go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path database/migrations -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Rolling back database migrations..."
	@which migrate > /dev/null || (echo "migrate CLI not found. Installing..."; go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path database/migrations -database "$(DATABASE_URL)" down

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/slack-bot cmd/api/main.go

run: docker-up migrate-up
	go run cmd/api/main.go

clean:
	rm -rf bin/ coverage.out
	go clean
