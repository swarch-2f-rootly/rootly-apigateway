# Rootly API Gateway

**Rootly API Gateway** is a GraphQL-based API Gateway built for the Rootly Smart Plant Monitoring System. It acts as the central entry point that aggregates and orchestrates communication between multiple microservices in the Rootly ecosystem, providing a unified GraphQL interface for frontend applications.

## 🌱 **About Rootly**

Rootly is a comprehensive IoT-based plant monitoring system designed to help users track and analyze environmental conditions for optimal plant care. The system consists of multiple specialized microservices:

- **Analytics Service**: Processes sensor data and generates insights
- **Data Management Service**: Handles sensor data ingestion and storage
- **User Plant Management Service**: Manages user profiles and plant configurations
- **Authentication & Roles Service**: Handles user authentication and authorization

The API Gateway serves as the orchestration layer, providing a single GraphQL endpoint that intelligently routes requests to the appropriate backend services, handles data aggregation, and ensures consistent API responses for client applications.

## 🏗️ **Architecture**

This API Gateway follows **Hexagonal Architecture** principles with clean separation of concerns:

- **Domain Layer**: Core business entities and value objects
- **Application Layer**: Use cases and business logic orchestration
- **Infrastructure Layer**: External service adapters (HTTP clients, GraphQL resolvers)
- **Ports & Adapters**: Interface definitions for external communication

## 🚀 **Features**

- **GraphQL Unified API**: Single endpoint for all microservice operations
- **Service Orchestration**: Intelligent routing to backend services
- **Real-time Analytics**: Sensor data processing and trend analysis
- **Health Monitoring**: Service health checks and system status
- **CORS Support**: Cross-origin resource sharing for web applications
- **GraphQL Playground**: Interactive API exploration and testing
- **Hexagonal Architecture**: Clean, maintainable, and testable codebase

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

### 🔍 **Basic Health Check**
Start with this simple query to verify the service is running:

```graphql
query {
  getSupportedMetrics
}
```

### 🩺 **Service Health Status**
Check the health of the analytics service:

```graphql
query {
  getAnalyticsHealth {
    status
    service
    influxdb
    influxdbUrl
    timestamp
  }
}
```

### 📊 **Single Metric Report**
Get analytics data for a specific metric and controller:

```graphql
query SingleMetricReport {
  getSingleMetricReport(
    metricName: "temperature"
    controllerId: "controller-123"
    filters: {
      startTime: "2024-01-01T00:00:00Z"
      endTime: "2024-01-31T23:59:59Z"
      limit: 100
    }
  ) {
    controllerId
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

### 🔢 **Multi-Metric Report**
Get data for multiple metrics and controllers:

```graphql
query MultiMetricReport {
  getMultiMetricReport(
    input: {
      controllers: ["controller-123", "controller-456"]
      metrics: ["temperature", "humidity", "light_intensity"]
      filters: {
        startTime: "2024-01-01T00:00:00Z"
        endTime: "2024-01-31T23:59:59Z"
        limit: 50
      }
    }
  ) {
    generatedAt
    totalControllers
    totalMetrics
    reports {
      controllerId
      dataPointsCount
      metrics {
        metricName
        value
        unit
      }
    }
  }
}
```

### 📈 **Trend Analysis**
Analyze trends over time for specific metrics:

```graphql
query TrendAnalysis {
  getTrendAnalysis(
    input: {
      metricName: "temperature"
      controllerId: "controller-123"
      startTime: "2024-01-01T00:00:00Z"
      endTime: "2024-01-07T23:59:59Z"
      interval: "1h"
    }
  ) {
    metricName
    controllerId
    interval
    generatedAt
    totalPoints
    averageValue
    minValue
    maxValue
    dataPoints {
      timestamp
      value
      interval
    }
  }
}
```

### 🧪 **Simple Test Queries**

**Quick Health Check:**
```graphql
{ getAnalyticsHealth { status } }
```

**Available Metrics:**
```graphql
{ getSupportedMetrics }
```

### 📝 **Testing Workflow**

1. **Start with supported metrics** to see what's available
2. **Check service health** to ensure backend connectivity
3. **Use real controller IDs** from your analytics database
4. **Start with small date ranges** for faster responses
5. **Gradually test more complex multi-metric queries**

## 🔧 **Configuration**

The service reads configuration from environment variables. Copy `.env.example` to `.env` and adjust as needed:

### **Backend Services**
```env
ANALYTICS_SERVICE_URL=http://localhost:8001
AUTH_SERVICE_URL=http://localhost:8002
DATA_MANAGEMENT_SERVICE_URL=http://localhost:8003
PLANT_MANAGEMENT_SERVICE_URL=http://localhost:8004
```

### **Server Configuration**
```env
PORT=8080                                # API Gateway port
GIN_MODE=debug                          # Gin framework mode (debug/release)
```

### **GraphQL Configuration**
```env
GRAPHQL_PLAYGROUND_ENABLED=true         # Enable GraphQL Playground UI
GRAPHQL_INTROSPECTION_ENABLED=true     # Enable GraphQL introspection
```

### **CORS Configuration**
```env
CORS_ALLOW_ALL_ORIGINS=true            # Allow all origins (development)
```

### **Logging Configuration**
```env
LOG_LEVEL=info                         # Logging level (debug/info/warn/error)
LOG_FORMAT=json                        # Log output format (json/text)
```

## 🎮 **GraphQL Playground**

When `GRAPHQL_PLAYGROUND_ENABLED=true`, access the interactive GraphQL IDE at:

```
http://localhost:8080/playground
```

**Features:**
- **Query Editor**: Write and execute GraphQL queries
- **Schema Explorer**: Browse available types and fields
- **Documentation**: Auto-generated API documentation
- **Query History**: Access previously executed queries
- **Variables Support**: Test queries with dynamic variables

**Getting Started in Playground:**
1. Open the Playground in your browser
2. Use the **Docs** panel to explore available queries
3. Start with simple queries like `{ getSupportedMetrics }`
4. Use the **Schema** tab to understand data structures
5. Copy and paste the example queries from this README

## Architecture

## 🏛️ **Architecture Overview**

This API Gateway implements **Hexagonal Architecture** (Ports and Adapters) for maximum maintainability and testability:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GraphQL       │    │   HTTP Client   │    │   Config        │
│   Resolvers     │    │   Adapters      │    │   Management    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
            ┌─────────────────────┴─────────────────────┐
            │              PORTS LAYER                  │
            │  ┌─────────────────┐ ┌─────────────────┐  │
            │  │ AnalyticsService│ │ AnalyticsClient │  │
            │  │   Interface     │ │   Interface     │  │
            │  └─────────────────┘ └─────────────────┘  │
            └─────────────────────┬─────────────────────┘
                                 │
            ┌─────────────────────┴─────────────────────┐
            │             DOMAIN LAYER                  │
            │  ┌─────────────────┐ ┌─────────────────┐  │
            │  │   Analytics     │ │     Domain      │  │
            │  │   Entities      │ │   Value Objects │  │
            │  └─────────────────┘ └─────────────────┘  │
            └───────────────────────────────────────────┘
```

### **Layer Responsibilities:**

- **🎯 Domain Layer**: Core business entities (`AnalyticsReport`, `MetricResult`, etc.)
- **🔌 Ports Layer**: Interface contracts (`AnalyticsService`, `AnalyticsClient`)
- **🔧 Adapters Layer**: External integrations (GraphQL resolvers, HTTP clients)
- **⚙️ Infrastructure**: Configuration, logging, and cross-cutting concerns

### **Benefits:**
- **Testability**: Easy to mock external dependencies
- **Maintainability**: Clear separation of concerns
- **Flexibility**: Easy to swap implementations
- **Domain-Driven**: Business logic stays in the core

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

## 🛠️ **Troubleshooting**

### **GraphQL Query Issues**

#### **"Field not found" Errors**
```bash
# Check the schema in GraphQL Playground
# Use the correct field names: controllerId (not controllerID)
```

#### **"Cannot query field" Errors**
```bash
# Verify field names match the schema exactly
# Check if using the correct input parameter structure
```

#### **Backend Connection Issues**
```bash
# Test backend service directly
curl http://localhost:8001/api/v1/analytics/health

# Check environment variables
echo $ANALYTICS_SERVICE_URL
```

### **Common Development Issues**

#### **GraphQL Code Generation**
```bash
make clean
make generate
```

#### **Dependency Issues**
```bash
make deps
```

#### **Development Tools Missing**
```bash
make install-tools
```

#### **Port Already in Use**
```bash
# Check what's using port 8080
lsof -i :8080

# Change port in .env file
echo "PORT=8081" >> .env
```

### **Backend Service Issues**

#### **Analytics Service Not Responding**
```bash
# Verify analytics service is running
curl http://localhost:8001/api/v1/analytics/health

# Check analytics service logs
```

#### **Invalid Controller IDs**
```bash
# Use getSupportedMetrics to see available data
# Check your database for actual controller IDs
```

#### **Date Range Issues**
```bash
# Use recent date ranges
# Verify your database has data for the specified timeframe
```

### **GraphQL Playground Issues**

#### **Playground Not Loading**
- Check `GRAPHQL_PLAYGROUND_ENABLED=true` in `.env`
- Verify you're accessing `http://localhost:8080/playground`
- Check browser console for errors

#### **Schema Not Loading**
- Restart the API Gateway service
- Check for GraphQL generation errors in logs

### **Getting Help**
```bash
make help         # Show all available commands
make info         # Show project information
```
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