package http

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/config"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/domain"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/services"
)

// GatewayHandler handles HTTP requests for the API Gateway
type GatewayHandler struct {
	gatewayService *services.GatewayService
	configProvider ports.ConfigProvider
	logger         ports.Logger
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(
	gatewayService *services.GatewayService,
	configProvider ports.ConfigProvider,
	logger ports.Logger,
) *GatewayHandler {
	return &GatewayHandler{
		gatewayService: gatewayService,
		configProvider: configProvider,
		logger:         logger,
	}
}

// HandleRequest handles incoming HTTP requests
func (gh *GatewayHandler) HandleRequest(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Build request context
	reqCtx := &domain.RequestContext{
		RequestID: requestID,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Headers:   make(map[string]string),
		Query:     make(map[string]string),
		StartTime: startTime,
	}

	// Extract headers
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			reqCtx.Headers[strings.ToLower(key)] = values[0]
		}
	}

	// Extract query parameters
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			reqCtx.Query[key] = values[0]
		}
	}

	// Extract body for POST/PUT requests
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		contentType := strings.ToLower(c.GetHeader("Content-Type"))

		// Only parse as JSON if content-type is application/json
		if strings.Contains(contentType, "application/json") {
			if c.Request.Body != nil {
				var body interface{}
				if err := c.ShouldBindJSON(&body); err == nil {
					reqCtx.Body = body
				}
			}
		} else if strings.Contains(contentType, "multipart/form-data") {
			// For multipart/form-data, store the raw body and mark it for special handling
			if c.Request.Body != nil {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err != nil {
					gh.logger.Error("Failed to read multipart body", err, map[string]interface{}{
						"request_id": requestID,
					})
				} else {
					reqCtx.Body = bodyBytes
					reqCtx.Headers["x-multipart-body"] = "true"
				}
			}
		} else {
			// For other content types, try to read the body as raw bytes if it's not JSON
			if c.Request.Body != nil && c.Request.ContentLength > 0 {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err != nil {
					gh.logger.Error("Failed to read raw body", err, map[string]interface{}{
						"request_id": requestID,
					})
				} else {
					reqCtx.Body = bodyBytes
				}
			}
		}
	}

	gh.logger.Info("Request received", map[string]interface{}{
		"request_id": requestID,
		"method":     reqCtx.Method,
		"path":       reqCtx.Path,
		"user_agent": c.GetHeader("User-Agent"),
		"remote_ip":  c.ClientIP(),
	})

	// Process request
	ctx := context.WithValue(context.Background(), "request_id", requestID)
	response, err := gh.gatewayService.ProcessRequest(ctx, reqCtx)

	// Handle errors
	if err != nil {
		gh.logger.Error("Request processing failed", err, map[string]interface{}{
			"request_id": requestID,
			"duration":   time.Since(startTime).Milliseconds(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": requestID,
		})
		return
	}

	// Set response headers
	if response.Headers != nil {
		for key, value := range response.Headers {
			c.Header(key, value)
		}
	}

	// Log successful response
	gh.logger.Info("Request completed", map[string]interface{}{
		"request_id":  requestID,
		"status_code": response.StatusCode,
		"duration":    time.Since(startTime).Milliseconds(),
	})

	// Send response
	c.JSON(response.StatusCode, response.Body)
}

// HandleHealth handles health check requests
func (gh *GatewayHandler) HandleHealth(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"services":  make(map[string]interface{}),
	}

	// Add basic service status (could be enhanced with actual health checks)
	services := []string{"analytics", "auth", "data_management", "plant_management"}
	for _, service := range services {
		if serviceInfo, exists := gh.configProvider.GetServiceConfig(service); exists {
			health["services"].(map[string]interface{})[service] = gin.H{
				"url":    serviceInfo.URL,
				"status": "unknown", // Could implement actual health checking
			}
		}
	}

	c.JSON(http.StatusOK, health)
}

// HandleMetrics handles metrics endpoint
func (gh *GatewayHandler) HandleMetrics(c *gin.Context) {
	metrics := gin.H{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"gateway": gin.H{
			"uptime":  time.Since(time.Now()).String(), // This should be actual uptime
			"version": "1.0.0",
		},
		"requests": gin.H{
			"total":   0, // Would be tracked by metrics collector
			"success": 0,
			"errors":  0,
		},
		"services": gin.H{
			"total":   len(gh.configProvider.(*ConfigProvider).config.Services),
			"healthy": 0, // Would be updated by health checks
		},
	}

	c.JSON(http.StatusOK, metrics)
}

// RegisterRoutes registers all gateway routes
func (gh *GatewayHandler) RegisterRoutes(router *gin.Engine) {
	// Operational endpoints (infrastructure, no versioning)
	// These are used by monitoring systems, load balancers, and orchestrators
	router.GET("/health", gh.HandleHealth)   // Gateway health check
	router.HEAD("/health", gh.HandleHealth)  // Gateway health check (HEAD)
	router.GET("/healthz", gh.HandleHealth)  // Kubernetes-style alias
	router.HEAD("/healthz", gh.HandleHealth) // Kubernetes-style alias (HEAD)
	router.GET("/metrics", gh.HandleMetrics) // Prometheus metrics

	// Business API routes (versioned, dynamic routing from config.yaml)
	// Pattern: /api/v1/* → processed by NoRoute handler
	router.NoRoute(gh.HandleRequest)
}

// ConfigProvider implements ports.ConfigProvider interface
type ConfigProvider struct {
	config *config.Config
	logger ports.Logger
}

// NewConfigProvider creates a new config provider
func NewConfigProvider(config *config.Config, logger ports.Logger) *ConfigProvider {
	return &ConfigProvider{
		config: config,
		logger: logger,
	}
}

// GetRouteConfig retrieves route configuration for a path and method
func (cp *ConfigProvider) GetRouteConfig(path string, method string) (*ports.RouteConfig, bool) {
	for _, route := range cp.config.Routes {
		if cp.matchRoute(route, path, method) {
			return &ports.RouteConfig{
				Path:         route.Path,
				Method:       route.Method,
				Mode:         route.Mode,
				Strategy:     route.Strategy,
				Upstream:     route.Upstream,
				TargetPath:   route.TargetPath,
				AuthRequired: route.AuthRequired,
				Upstreams:    cp.convertUpstreams(route.Upstreams),
				Metadata:     route.Metadata,
			}, true
		}
	}
	return nil, false
}

// GetServiceConfig retrieves service configuration by name
func (cp *ConfigProvider) GetServiceConfig(serviceName string) (*ports.ServiceInfo, bool) {
	if service, exists := cp.config.Services[serviceName]; exists {
		return &ports.ServiceInfo{
			Name:    serviceName,
			URL:     service.URL,
			Timeout: service.Timeout.String(),
		}, true
	}
	return nil, false
}

// GetStrategyConfig retrieves strategy configuration by name
func (cp *ConfigProvider) GetStrategyConfig(strategyName string) (map[string]interface{}, bool) {
	if strategy, exists := cp.config.Strategies[strategyName]; exists {
		result := make(map[string]interface{})
		result["timeout"] = strategy.Timeout.String()
		result["parallel_requests"] = strategy.ParallelRequests
		result["failure_policy"] = strategy.FailurePolicy
		result["introspection_enabled"] = strategy.IntrospectionEnabled
		result["playground_enabled"] = strategy.PlaygroundEnabled
		result["preserve_headers"] = strategy.PreserveHeaders
		result["proxy_timeout"] = strategy.ProxyTimeout.String()
		return result, true
	}
	return nil, false
}

// ReloadConfig reloads the configuration
func (cp *ConfigProvider) ReloadConfig() error {
	newConfig := config.LoadConfig()
	cp.config = newConfig
	cp.logger.Info("Configuration reloaded", map[string]interface{}{
		"routes_count":     len(newConfig.Routes),
		"services_count":   len(newConfig.Services),
		"strategies_count": len(newConfig.Strategies),
	})
	return nil
}

// matchRoute checks if a route matches the given path and method
func (cp *ConfigProvider) matchRoute(route config.RouteConfig, path string, method string) bool {
	// Check method first
	if route.Method != method && route.Method != "*" {
		return false
	}

	// Simple path matching with wildcard support
	routeParts := strings.Split(strings.Trim(route.Path, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	cp.logger.Debug("Matching route", map[string]interface{}{
		"route_path":   route.Path,
		"request_path": path,
		"route_parts":  routeParts,
		"path_parts":   pathParts,
	})

	// If route ends with wildcard (*), it should match any path that starts with the route prefix
	if len(routeParts) > 0 && routeParts[len(routeParts)-1] == "*" {
		// Remove the wildcard from route parts for comparison
		routePrefix := routeParts[:len(routeParts)-1]

		// Path must have at least as many parts as the route prefix (can be equal or more)
		if len(pathParts) < len(routePrefix) {
			return false
		}

		// Check that all prefix parts match
		for i, routePart := range routePrefix {
			if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
				// This is a path parameter, skip validation
				continue
			}
			if routePart != pathParts[i] {
				return false
			}
		}

		return true
	}

	// Exact match for non-wildcard routes
	if len(routeParts) != len(pathParts) {
		return false
	}

	for i, routePart := range routeParts {
		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			// This is a path parameter, skip validation
			continue
		}
		if routePart != pathParts[i] {
			return false
		}
	}

	return true
}

// convertUpstreams converts config upstreams to ports upstreams
func (cp *ConfigProvider) convertUpstreams(upstreams []config.UpstreamConfig) []ports.UpstreamConfig {
	result := make([]ports.UpstreamConfig, len(upstreams))
	for i, upstream := range upstreams {
		result[i] = ports.UpstreamConfig{
			Service:  upstream.Service,
			Endpoint: upstream.Endpoint,
			Method:   upstream.Method,
		}
	}
	return result
}
