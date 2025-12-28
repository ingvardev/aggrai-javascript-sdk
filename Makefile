.PHONY: help dev infra stop migrate sqlc gql test lint build

# Default target
help:
	@echo "AI Aggregator - Available commands:"
	@echo ""
	@echo "  make dev       - Start all services for development"
	@echo "  make infra     - Start infrastructure only (postgres, redis)"
	@echo "  make stop      - Stop all services"
	@echo "  make migrate   - Run database migrations"
	@echo "  make sqlc      - Generate sqlc code"
	@echo "  make gql       - Generate GraphQL code"
	@echo "  make test      - Run all tests"
	@echo "  make lint      - Run linters"
	@echo "  make build     - Build all services"
	@echo ""

# Development
dev: infra
	@echo "Starting development environment..."
	@cd apps/api && go run ./cmd/server &
	@cd apps/worker && go run ./cmd/worker &
	@cd apps/web && npm run dev

infra:
	@echo "Starting infrastructure..."
	docker-compose up -d postgres redis

stop:
	@echo "Stopping all services..."
	docker-compose down

# Database
migrate:
	@echo "Running migrations..."
	cd infrastructure/postgres && golang-migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Rolling back migrations..."
	cd infrastructure/postgres && golang-migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@echo "Creating new migration..."
	@read -p "Migration name: " name; \
	cd infrastructure/postgres && golang-migrate create -ext sql -dir migrations -seq $$name

# Code generation
sqlc:
	@echo "Generating sqlc code..."
	cd infrastructure/postgres && sqlc generate

gql:
	@echo "Generating GraphQL code..."
	cd apps/api && go generate ./...

# Testing
test:
	@echo "Running tests..."
	go test ./... -v

test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Linting
lint:
	@echo "Running linters..."
	golangci-lint run ./...
	cd apps/web && npm run lint

# Building
build: build-api build-worker build-web

build-api:
	@echo "Building API..."
	cd apps/api && go build -o ../../bin/api ./cmd/server

build-worker:
	@echo "Building Worker..."
	cd apps/worker && go build -o ../../bin/worker ./cmd/worker

build-web:
	@echo "Building Web..."
	cd apps/web && npm run build

# Docker
docker-build:
	@echo "Building Docker images..."
	docker-compose build api worker web

docker-up:
	@echo "Starting all Docker services..."
	docker-compose --profile app up -d

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf apps/web/.next
	rm -rf apps/web/node_modules
