# Makefile for UIM Go Server
# Copyright 2025 convexwf

.PHONY: help build run test clean docker-build docker-up docker-down docker-logs init-db migrate

# Variables
APP_NAME := uim-server
BINARY_PATH := bin/$(APP_NAME)
DOCKER_IMAGE := uim-go:latest
DOCKER_CONTAINER := uim-server

# Default target
help:
	@echo "Available targets:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run all tests"
	@echo "  make test-cover   - Run tests with coverage"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start Docker services (PostgreSQL + Redis)"
	@echo "  make docker-down  - Stop Docker services"
	@echo "  make docker-logs  - View Docker logs"
	@echo "  make init-db      - Run init_db.sh (or via docker if postgres is up and psql missing)"
	@echo "  make migrate      - (deprecated) Use init-db; schema is SQL-first"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BINARY_PATH) ./cmd/server
	@echo "Build complete: $(BINARY_PATH)"

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/server

# Run all tests
test:
	@echo "Running tests..."
	@go test ./internal/... -v

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@go test ./internal/... -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

# Start Docker services (PostgreSQL + Redis + UIM Server)
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "Docker services started (postgres, redis, uim-server)"

# Stop Docker services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "Docker services stopped"

# View Docker logs
docker-logs:
	@docker-compose logs -f

# Initialize DB schema (SQL-first). Run before first start or after schema changes.
init-db:
	@if command -v psql >/dev/null 2>&1; then \
		./scripts/init_db.sh; \
	elif docker ps --format '{{.Names}}' | grep -q 'uim-postgres'; then \
		echo "psql not found; running migration inside uim-postgres container..."; \
		docker exec -i uim-postgres psql -U $${POSTGRES_USER:-uim_user} -d $${POSTGRES_DB:-uim_db} -v ON_ERROR_STOP=1 < migrations/000001_initial_schema.up.sql; \
		echo "Done."; \
	else \
		echo "Error: install psql and run ./scripts/init_db.sh, or start postgres and run make init-db again"; exit 1; \
	fi

# Deprecated: schema is created by init-db, not on server startup
migrate:
	@echo "Run 'make init-db' before starting the server. See doc/feature/database-migrations.md"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./... || echo "golangci-lint not installed, skipping..."

# Run all checks (format, lint, test)
check: fmt lint test
	@echo "All checks passed"
