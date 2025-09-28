# Rootly API Gateway

GraphQL API Gateway for the Rootly ecosystem that aggregates analytics and other microservices.

## Prerequisites

- Go 1.21 or higher
- Make (for using Makefile commands)
- Docker (optional, for containerized deployment)

## Installation

Clone the repository and use the quickstart command:

```bash
git clone https://github.com/swarch-2f-rootly/rootly-apigateway.git
cd rootly-apigateway
make quickstart
```

Or install manually:

```bash
make install-tools  # Install development tools
make setup-dev      # Setup development environment
```

## Quick Start

### For New Developers

```bash
# Complete setup for new developers
make quickstart
```

### Manual Setup

1. Copy environment variables:

```bash
cp .env.example .env
```

2. Install dependencies and generate code:

```bash
make deps
make generate
```

3. Start the analytics backend service (prerequisite):

```bash
# Make sure the analytics service is running on localhost:8001
```

4. Run the API Gateway:

```bash
make dev
```

5. Access the GraphQL Playground:

```
http://localhost:8080/playground
```

## Available Endpoints

- **GraphQL Endpoint**: `POST/GET /graphql`
- **GraphQL Playground**: `GET /playground` (if enabled)
- **Health Check**: `GET /health`

## Example GraphQL Queries

### Get Supported Metrics

```graphql
query {
    getSupportedMetrics
}
```

### Get Single Metric Report

```graphql
query {
    getSingleMetricReport(
        metricName: "temperature"
        controllerID: "controller-123"
        filters: {
            startTime: "2024-01-01T00:00:00Z"
            endTime: "2024-01-02T00:00:00Z"
            limit: 100
        }
    ) {
        controllerID
        generatedAt
        dataPointsCount
        metrics {
            metricName
            value
            unit
            calculatedAt
        }
    }
}
```

### Get Analytics Health

```graphql
query {
    getAnalyticsHealth {
        serviceName
        status
        checkedAt
        version
        dependencies
    }
}
```

## Configuration

The service reads configuration from environment variables (see `.env.example`):

- `ANALYTICS_SERVICE_URL`: URL of the analytics backend service
- `PORT`: Port for the API Gateway (default: 8080)
- `GRAPHQL_PLAYGROUND_ENABLED`: Enable/disable GraphQL Playground
- `CORS_ALLOW_ALL_ORIGINS`: Enable CORS for all origins

## Architecture

This API Gateway follows hexagonal architecture principles:

- **Domain**: Core business entities and value objects
- **Ports**: Interfaces for external communication
- **Adapters**: HTTP clients and GraphQL resolvers
- **Services**: Business logic layer

## Development

### Makefile Commands

This project includes a comprehensive Makefile for development and deployment tasks:

#### Development Commands

```bash
make dev          # Run in development mode with auto-reload
make build        # Build the binary
make run          # Build and run the binary
make generate     # Generate GraphQL code
```

#### Testing and Quality

```bash
make test         # Run all tests
make test-coverage # Run tests with coverage report
make lint         # Run linter (requires golangci-lint)
make fmt          # Format code
make vet          # Run go vet
make check        # Run all quality checks (fmt, vet, lint, test)
```

#### Dependency Management

```bash
make deps         # Download and tidy dependencies
make install      # Install dependencies
make install-tools # Install development tools
```

#### Docker Commands

```bash
make docker-build # Build Docker image
make docker-run   # Build and run Docker container
```

#### Utilities

```bash
make clean        # Clean build artifacts
make setup-dev    # Setup development environment
make quickstart   # Complete setup for new developers
make info         # Show project information
make help         # Show all available commands
```

#### Production

```bash
make build-prod   # Build optimized binary for production
```

### Common Workflows

#### Daily Development

```bash
make dev          # Start development server
```

#### Before Committing

```bash
make check        # Run all quality checks
```

#### First Time Setup

```bash
make quickstart   # Complete automated setup
```

### Generate GraphQL Code

```bash
make generate
```

### Project Structure

```
cmd/server/           # Application entry point
internal/
  adapters/
    graphql/          # GraphQL resolvers and generated code
    http/             # HTTP clients for backend services
  config/             # Configuration management
  core/
    ports/            # Interface definitions
    services/         # Business logic
  domain/             # Domain entities and value objects
```

## Troubleshooting

### Common Issues

#### GraphQL Generation Errors

```bash
make clean
make generate
```

#### Dependency Issues

```bash
make deps
```

#### Development Tools Missing

```bash
make install-tools
```

#### Port Already in Use

Check if another service is running on port 8080:

```bash
lsof -i :8080
```

Change the port in your `.env` file:

```
PORT=8081
```

### Getting Help

```bash
make help    # Show all available commands
make info    # Show project information
```