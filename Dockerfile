# Rootly API Gateway Dockerfile

# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (required for go mod and HTTPS)
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=1.0.0 -X main.buildTime=$(date -u +%Y%m%d-%H%M%S)" \
    -o rootly-apigateway \
    cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Add ca-certificates for SSL/TLS and create non-root user
RUN apk --no-cache add ca-certificates wget && \
    addgroup -g 1001 appgroup && \
    adduser -D -s /bin/sh -u 1001 -G appgroup appuser

# Create app directory and set ownership
WORKDIR /app
RUN chown appuser:appgroup /app

# Copy the binary and config from builder stage
COPY --from=builder /app/rootly-apigateway .
COPY --from=builder /app/config.yml .

# Set ownership and permissions
RUN chown appuser:appgroup rootly-apigateway config.yml && \
    chmod +x rootly-apigateway

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV CONFIG_FILE=/app/config.yaml

# Run the binary
CMD ["./rootly-apigateway"]
