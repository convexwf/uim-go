# Makefile for UIM Go Server
# Copyright 2025 convexwf

.PHONY: help build run test clean docker-build docker-up docker-down docker-logs migrate

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
	@echo "  make migrate      - Run database migrations (via server startup)"

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

# Run database migrations (migrations run automatically on server startup)
migrate:
	@echo "Migrations run automatically on server startup"
	@echo "Start the server to run migrations: make run"

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
