# Makefile for Rootly API Gateway
.PHONY: help build run test test-unit test-integration test-docker test-env-start test-env-stop clean install dev docker-build docker-run docker-compose-up docker-compose-down generate lint fmt vet deps check coverage info quickstart

# Variables
BINARY_NAME=rootly-apigateway
BUILD_DIR=build
MAIN_PATH=cmd/server/main.go
DOCKER_IMAGE=rootly-apigateway
DOCKER_TAG=latest
VERSION?=1.0.0
BUILD_TIME=$(shell date -u +%Y%m%d-%H%M%S)
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Color output
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Default target
help:
	@echo "$(BLUE)Rootly API Gateway - Makefile Commands$(NC)"
	@echo ""
	@echo "$(YELLOW)Build & Run:$(NC)"
	@echo "  build              - Build the binary for Linux (production)"
	@echo "  build-local        - Build the binary for current platform"
	@echo "  run                - Build and run the application"
	@echo "  dev                - Run in development mode (live reload)"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  test               - Run all tests (unit + integration)"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-integration   - Run integration tests only"
	@echo "  test-docker        - Run full test suite with Docker"
	@echo "  test-env-start     - Start test environment (manual testing)"
	@echo "  test-env-stop      - Stop test environment"
	@echo "  coverage           - Generate test coverage report"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  deps               - Download and tidy dependencies"
	@echo "  generate           - Generate GraphQL code"
	@echo "  fmt                - Format code"
	@echo "  lint               - Run linter (requires golangci-lint)"
	@echo "  vet                - Run go vet"
	@echo "  check              - Run all quality checks"
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  docker-compose-up  - Start all services (full platform)"
	@echo "  docker-compose-down- Stop all services"
	@echo ""
	@echo "$(YELLOW)Quick Start:$(NC)"
	@echo "  quickstart         - Setup everything for new developers"
	@echo "  info               - Show project information"
	@echo "  clean              - Clean build artifacts"
	@echo ""

# Build the binary
build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built successfully: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for current platform
build-local: deps
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built successfully: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: build-local
	@echo "Starting $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run in development mode
dev: deps
	@echo "Starting development server..."
	go run $(MAIN_PATH)

# Run all tests
test: test-unit test-integration
	@echo "$(GREEN)All tests completed successfully$(NC)"

# Run unit tests
test-unit: deps
	@echo "$(YELLOW)Running unit tests...$(NC)"
	@go test -v -race ./internal/... ./test/unit/... || (echo "$(RED)Unit tests failed$(NC)" && exit 1)
	@echo "$(GREEN)Unit tests passed$(NC)"

# Run integration tests (requires services)
test-integration: deps
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@go test -v -tags=integration ./test/integration/... || (echo "$(RED)Integration tests failed$(NC)" && exit 1)
	@echo "$(GREEN)Integration tests passed$(NC)"

# Run full test suite with Docker (recommended)
test-docker:
	@echo "$(BLUE)Running full test suite with Docker...$(NC)"
	@./scripts/test-with-docker.sh

# Start test environment for manual testing
test-env-start:
	@echo "$(BLUE)Starting test environment...$(NC)"
	@./scripts/start_test_env.sh

# Stop test environment
test-env-stop:
	@echo "$(YELLOW)Stopping test environment...$(NC)"
	@./scripts/stop_test_env.sh

# Generate test coverage report
coverage: deps
	@echo "$(YELLOW)Generating coverage report...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./... 2>/dev/null || true
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -n 5
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html integration.coverage.out
	@go clean
	@echo "$(GREEN)Clean completed$(NC)"

# Install dependencies
install: deps

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Generate GraphQL code
generate:
	@echo "Generating GraphQL code..."
	cd internal/adapters/graphql && go run github.com/99designs/gqlgen generate
	@echo "GraphQL code generated successfully"

# Run linter (requires golangci-lint)
lint:
	@echo "$(YELLOW)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run && echo "$(GREEN)Lint passed$(NC)"; \
	else \
		echo "$(RED)golangci-lint not found. Install: make install-tools$(NC)"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "$(YELLOW)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

# Run go vet
vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	@go vet ./... && echo "$(GREEN)Vet passed$(NC)"

# Run all quality checks
check: fmt vet lint test
	@echo "$(GREEN)All quality checks passed!$(NC)"

# Build for production
build-prod: deps
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Production binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run Docker container
docker-run: docker-build
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Start services for development (requires docker-compose)
dev-services:
	@echo "Starting development services..."
	@if [ -f docker-compose.dev.yml ]; then \
		docker-compose -f docker-compose.dev.yml up -d; \
	else \
		echo "docker-compose.dev.yml not found"; \
	fi

# Stop development services
dev-services-stop:
	@echo "Stopping development services..."
	@if [ -f docker-compose.dev.yml ]; then \
		docker-compose -f docker-compose.dev.yml down; \
	else \
		echo "docker-compose.dev.yml not found"; \
	fi

# Setup development environment
setup-dev: deps generate
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created from .env.example"; \
	fi
	@echo "Development environment ready!"
	@echo "Run 'make dev' to start the development server"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/99designs/gqlgen@latest

# Docker Compose commands
docker-compose-up:
	@echo "$(BLUE)Starting all services with Docker Compose...$(NC)"
	@if [ -d "../rootly-deploy" ]; then \
		(cd ../rootly-deploy && docker-compose up -d --build) && \
		echo "$(GREEN)Services started. Gateway: http://localhost:8080$(NC)"; \
	else \
		echo "$(RED)Error: rootly-deploy directory not found$(NC)"; \
		exit 1; \
	fi

docker-compose-down:
	@echo "$(YELLOW)Stopping all services...$(NC)"
	@if [ -d "../rootly-deploy" ]; then \
		(cd ../rootly-deploy && docker-compose down -v) && \
		echo "$(GREEN)Services stopped$(NC)"; \
	else \
		echo "$(RED)Error: rootly-deploy directory not found$(NC)"; \
		exit 1; \
	fi

docker-compose-logs:
	@echo "$(BLUE)Showing service logs...$(NC)"
	@if [ -d "../rootly-deploy" ]; then \
		(cd ../rootly-deploy && docker-compose logs -f); \
	else \
		echo "$(RED)Error: rootly-deploy directory not found$(NC)"; \
		exit 1; \
	fi

# Quick development setup
quick-start: deps docker-compose-up
	@sleep 5
	@echo ""
	@echo "$(GREEN)=======================================$(NC)"
	@echo "$(GREEN)Quick Start Completed!$(NC)"
	@echo "$(GREEN)=======================================$(NC)"
	@echo ""
	@echo "$(BLUE)Available Endpoints:$(NC)"
	@echo "  Gateway Health:  http://localhost:8080/health"
	@echo "  Gateway Metrics: http://localhost:8080/metrics"
	@echo "  Dashboard:       http://localhost:8080/api/v1/dashboard"
	@echo ""
	@echo "$(YELLOW)Quick Test Commands:$(NC)"
	@echo "  curl http://localhost:8080/health | jq '.'"
	@echo "  curl http://localhost:8080/api/v1/dashboard | jq '.'"
	@echo ""
	@echo "$(BLUE)Useful Commands:$(NC)"
	@echo "  make test-docker    - Run all tests"
	@echo "  make test-env-start - Start test environment"
	@echo "  make logs          - View logs"
	@echo ""

# Show project info
info:
	@echo "$(BLUE)=======================================$(NC)"
	@echo "$(BLUE)Rootly API Gateway - Info$(NC)"
	@echo "$(BLUE)=======================================$(NC)"
	@echo ""
	@echo "Project:       $(BINARY_NAME)"
	@echo "Version:       $(VERSION)"
	@echo "Go version:    $(shell go version)"
	@echo "Build dir:     $(BUILD_DIR)"
	@echo "Main path:     $(MAIN_PATH)"
	@echo "Docker image:  $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo ""
	@echo "$(YELLOW)Endpoints Structure:$(NC)"
	@echo "  Operational (no versioning):"
	@echo "    • GET /health         - Gateway health check"
	@echo "    • GET /healthz        - Kubernetes-style health"
	@echo "    • GET /metrics        - Prometheus metrics"
	@echo ""
	@echo "  Business API (versioned: /api/v1/):"
	@echo "    • GET /api/v1/health                  - Backend health proxy"
	@echo "    • GET /api/v1/users/{id}             - User lookup"
	@echo "    • GET /api/v1/payments/{id}          - Payment lookup (auth)"
	@echo "    • GET /api/v1/dashboard              - Dashboard orchestration"
	@echo "    • GET /api/v1/plant/{id}/full-report - Plant report orchestration"
	@echo "    • POST /graphql                      - GraphQL local schema"
	@echo "    • POST /graphql-proxy                - GraphQL proxy"
	@echo ""

# Quick start for new developers
quickstart: install-tools setup-dev
	@echo ""
	@echo "$(GREEN)=======================================$(NC)"
	@echo "$(GREEN)Quickstart Completed!$(NC)"
	@echo "$(GREEN)=======================================$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "1. Configure your .env file if needed"
	@echo "2. Run '$(YELLOW)make docker-compose-up$(NC)' to start all services"
	@echo "3. Run '$(YELLOW)make dev$(NC)' to start the API Gateway in dev mode"
	@echo "4. Visit http://localhost:8080/health to test"
	@echo ""
	@echo "$(YELLOW)Useful commands:$(NC)"
	@echo "  make help       - Show all available commands"
	@echo "  make test       - Run tests"
	@echo "  make info       - Show project information"
	@echo ""

# Alias for docker-compose-logs
logs: docker-compose-logs