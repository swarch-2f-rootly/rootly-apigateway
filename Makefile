# Makefile for Rootly API Gateway
.PHONY: help build run test clean install dev docker-build docker-run generate lint fmt vet deps

# Variables
BINARY_NAME=rootly-apigateway
BUILD_DIR=build
MAIN_PATH=cmd/server/main.go
DOCKER_IMAGE=rootly-apigateway
DOCKER_TAG=latest

# Default target
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  run         - Run the application"
	@echo "  dev         - Run in development mode with auto-reload"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  install     - Install dependencies"
	@echo "  generate    - Generate GraphQL code"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  docker-build- Build Docker image"
	@echo "  docker-run  - Run Docker container"

# Build the binary
build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run in development mode
dev: deps
	@echo "Starting development server..."
	go run $(MAIN_PATH)

# Run tests
test: deps
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage: deps
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean

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
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run all quality checks
check: fmt vet lint test
	@echo "All quality checks completed"

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
	@echo "Development tools installed"

# Show project info
info:
	@echo "Project: $(BINARY_NAME)"
	@echo "Go version: $(shell go version)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "Main path: $(MAIN_PATH)"
	@echo "Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Quick start for new developers
quickstart: install-tools setup-dev
	@echo ""
	@echo "Quick start completed!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Configure your .env file if needed"
	@echo "2. Start the analytics backend service"
	@echo "3. Run 'make dev' to start the API Gateway"
	@echo "4. Visit http://localhost:8080/playground to test GraphQL"