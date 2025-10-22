package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/domain"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
)

// GatewayService implements the main gateway orchestration logic
type GatewayService struct {
	strategyManager     ports.StrategyManager
	serviceOrchestrator ports.ServiceOrchestrator
	authService         ports.AuthService
	logger              ports.Logger
	httpClient          ports.HTTPClient
	configProvider      ports.ConfigProvider
}

// NewGatewayService creates a new gateway service
func NewGatewayService(
	strategyManager ports.StrategyManager,
	serviceOrchestrator ports.ServiceOrchestrator,
	authService ports.AuthService,
	logger ports.Logger,
	httpClient ports.HTTPClient,
	configProvider ports.ConfigProvider,
) *GatewayService {
	return &GatewayService{
		strategyManager:     strategyManager,
		serviceOrchestrator: serviceOrchestrator,
		authService:         authService,
		logger:              logger,
		httpClient:          httpClient,
		configProvider:      configProvider,
	}
}

// ProcessRequest processes an incoming request based on the route configuration
func (gs *GatewayService) ProcessRequest(ctx context.Context, reqCtx *domain.RequestContext) (*domain.Response, error) {
	// Find matching route
	routeConfig, found := gs.configProvider.GetRouteConfig(reqCtx.Path, reqCtx.Method)
	if !found {
		return &domain.Response{
			StatusCode: http.StatusNotFound,
			Body:       map[string]string{"error": "Route not found"},
		}, nil
	}

	// Convert to domain route
	route := &domain.Route{
		Path:         routeConfig.Path,
		Method:       routeConfig.Method,
		Mode:         domain.RouteMode(routeConfig.Mode),
		Strategy:     routeConfig.Strategy,
		Upstream:     routeConfig.Upstream,
		TargetPath:   routeConfig.TargetPath,
		AuthRequired: routeConfig.AuthRequired,
	}

	reqCtx.Route = route

	gs.logger.Info("Processing request", map[string]interface{}{
		"request_id": reqCtx.RequestID,
		"method":     reqCtx.Method,
		"path":       reqCtx.Path,
		"mode":       string(route.Mode),
		"strategy":   route.Strategy,
	})

	// Handle authentication if required
	if route.AuthRequired {
		user, err := gs.authenticateRequest(ctx, reqCtx)
		if err != nil {
			gs.logger.Error("Authentication failed", err, map[string]interface{}{
				"request_id": reqCtx.RequestID,
			})
			return &domain.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       map[string]string{"error": "Authentication failed"},
			}, nil
		}
		reqCtx.User = user
	}

	// Route based on mode
	switch route.Mode {
	case domain.ProxyMode:
		return gs.handleProxyMode(ctx, reqCtx, *routeConfig)
	case domain.LogicMode:
		return gs.handleLogicMode(ctx, reqCtx, *routeConfig)
	case domain.GraphQLMode:
		return gs.handleGraphQLMode(ctx, reqCtx, *routeConfig)
	default:
		return &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body:       map[string]string{"error": fmt.Sprintf("Unsupported route mode: %s", route.Mode)},
		}, nil
	}
}

// authenticateRequest handles request authentication
func (gs *GatewayService) authenticateRequest(ctx context.Context, reqCtx *domain.RequestContext) (*domain.User, error) {
	gs.logger.Info("🔐 Authenticating request", map[string]interface{}{
		"request_id":    reqCtx.RequestID,
		"method":        reqCtx.Method,
		"path":          reqCtx.Path,
		"headers_count": len(reqCtx.Headers),
	})

	// Log all headers for debugging
	for key, value := range reqCtx.Headers {
		if strings.ToLower(key) == "authorization" {
			// Mask the token for security
			maskedValue := value
			if len(value) > 20 {
				maskedValue = value[:20] + "..."
			}
			gs.logger.Info("🔑 Authorization header found", map[string]interface{}{
				"request_id":           reqCtx.RequestID,
				"header_key":           key,
				"header_value_preview": maskedValue,
				"value_length":         len(value),
			})
		} else {
			gs.logger.Debug("📋 Header received", map[string]interface{}{
				"request_id":   reqCtx.RequestID,
				"header_key":   key,
				"header_value": value,
			})
		}
	}

	// Check for API key first
	if apiKey, exists := reqCtx.Headers["x-api-key"]; exists {
		gs.logger.Debug("API key found, validating", map[string]interface{}{
			"request_id":     reqCtx.RequestID,
			"api_key_prefix": apiKey[:min(8, len(apiKey))],
		})
		if valid, err := gs.authService.ValidateAPIKey(ctx, apiKey); err != nil {
			return nil, err
		} else if valid {
			// Return a basic user info for API key auth
			return &domain.User{
				ID:       "api-key-user",
				Username: "api-key",
				Roles:    []string{"api-user"},
			}, nil
		}
	}

	// Check for JWT token
	if authHeader, exists := reqCtx.Headers["authorization"]; exists {
		gs.logger.Debug("Authorization header found", map[string]interface{}{
			"request_id":    reqCtx.RequestID,
			"header_prefix": authHeader[:min(20, len(authHeader))],
		})
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			userInfo, err := gs.authService.ValidateJWT(ctx, token)
			if err != nil {
				return nil, err
			}
			return &domain.User{
				ID:       userInfo.ID,
				Username: userInfo.Username,
				Email:    userInfo.Email,
				Roles:    userInfo.Roles,
				Metadata: userInfo.Metadata,
			}, nil
		}
	}

	gs.logger.Debug("No valid authentication provided", map[string]interface{}{
		"request_id":    reqCtx.RequestID,
		"headers_count": len(reqCtx.Headers),
	})
	return nil, fmt.Errorf("no valid authentication provided")
}

// handleProxyMode handles proxy mode requests
func (gs *GatewayService) handleProxyMode(ctx context.Context, reqCtx *domain.RequestContext, routeConfig ports.RouteConfig) (*domain.Response, error) {
	gs.logger.Info("🔀 Handling proxy mode", map[string]interface{}{
		"request_id":  reqCtx.RequestID,
		"upstream":    routeConfig.Upstream,
		"target_path": routeConfig.TargetPath,
		"route_path":  routeConfig.Path,
	})

	serviceInfo, found := gs.configProvider.GetServiceConfig(routeConfig.Upstream)
	if !found {
		gs.logger.Error("Upstream service not found", nil, map[string]interface{}{
			"request_id": reqCtx.RequestID,
			"upstream":   routeConfig.Upstream,
		})
		return &domain.Response{
			StatusCode: http.StatusBadGateway,
			Body:       map[string]string{"error": "Upstream service not configured"},
		}, nil
	}

	gs.logger.Info("📍 Service info found", map[string]interface{}{
		"request_id":   reqCtx.RequestID,
		"service_url":  serviceInfo.URL,
		"service_name": serviceInfo.Name,
	})

	// Create HTTP request from context
	httpRequest := gs.createHTTPRequestFromContext(reqCtx)

	// Determine strategy name (use default if not specified)
	strategyName := routeConfig.Strategy
	if strategyName == "" {
		strategyName = "proxy"
	}

	gs.logger.Info("🎯 Executing strategy", map[string]interface{}{
		"request_id":    reqCtx.RequestID,
		"strategy_name": strategyName,
	})

	// Execute strategy
	strategyParams := ports.StrategyParams{
		Request:     httpRequest,
		RouteConfig: routeConfig,
		Services: map[string]ports.ServiceInfo{
			routeConfig.Upstream: *serviceInfo,
		},
		UserInfo:   gs.convertUser(reqCtx.User),
		HTTPClient: gs.httpClient,
		Logger:     gs.logger,
	}

	result, err := gs.strategyManager.ExecuteStrategy(ctx, strategyName, strategyParams)
	if err != nil {
		gs.logger.Error("Proxy strategy execution failed", err, map[string]interface{}{
			"request_id": reqCtx.RequestID,
			"upstream":   routeConfig.Upstream,
		})
		return &domain.Response{
			StatusCode: http.StatusBadGateway,
			Body:       map[string]string{"error": "Upstream service error"},
		}, nil
	}

	gs.logger.Info("✅ Strategy executed successfully", map[string]interface{}{
		"request_id":  reqCtx.RequestID,
		"result_type": fmt.Sprintf("%T", result),
	})

	if httpResp, ok := result.(*http.Response); ok {
		resp, convertErr := gs.convertHTTPResponse(httpResp)
		gs.logger.Info("📦 Response converted", map[string]interface{}{
			"request_id":  reqCtx.RequestID,
			"status_code": resp.StatusCode,
			"has_error":   convertErr != nil,
		})
		return resp, convertErr
	}

	gs.logger.Info("📤 Returning direct result", map[string]interface{}{
		"request_id": reqCtx.RequestID,
	})

	return &domain.Response{
		StatusCode: http.StatusOK,
		Body:       result,
	}, nil
}

// handleLogicMode handles logic mode requests
func (gs *GatewayService) handleLogicMode(ctx context.Context, reqCtx *domain.RequestContext, routeConfig ports.RouteConfig) (*domain.Response, error) {
	// Collect service information for all upstreams
	services := make(map[string]ports.ServiceInfo)
	for _, upstream := range routeConfig.Upstreams {
		serviceInfo, found := gs.configProvider.GetServiceConfig(upstream.Service)
		if !found {
			gs.logger.Warn("Upstream service not configured", map[string]interface{}{
				"service": upstream.Service,
			})
			continue
		}
		services[upstream.Service] = *serviceInfo
	}

	// Create HTTP request from context
	httpRequest := gs.createHTTPRequestFromContext(reqCtx)

	// Execute logic strategy
	strategyParams := ports.StrategyParams{
		Request:     httpRequest,
		RouteConfig: routeConfig,
		Services:    services,
		UserInfo:    gs.convertUser(reqCtx.User),
		HTTPClient:  gs.httpClient,
		Logger:      gs.logger,
	}

	result, err := gs.strategyManager.ExecuteStrategy(ctx, routeConfig.Strategy, strategyParams)
	if err != nil {
		gs.logger.Error("Logic strategy execution failed", err, map[string]interface{}{
			"request_id": reqCtx.RequestID,
			"strategy":   routeConfig.Strategy,
		})
		return &domain.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       map[string]string{"error": "Strategy execution failed"},
		}, nil
	}

	return &domain.Response{
		StatusCode: http.StatusOK,
		Body:       result,
	}, nil
}

// handleGraphQLMode handles GraphQL mode requests
func (gs *GatewayService) handleGraphQLMode(ctx context.Context, reqCtx *domain.RequestContext, routeConfig ports.RouteConfig) (*domain.Response, error) {
	// Collect service information
	services := make(map[string]ports.ServiceInfo)
	if routeConfig.Upstream != "" {
		serviceInfo, found := gs.configProvider.GetServiceConfig(routeConfig.Upstream)
		if found {
			services[routeConfig.Upstream] = *serviceInfo
		}
	}

	// Create HTTP request from context
	httpRequest := gs.createHTTPRequestFromContext(reqCtx)

	// Execute GraphQL strategy
	strategyParams := ports.StrategyParams{
		Request:     httpRequest,
		RouteConfig: routeConfig,
		Services:    services,
		UserInfo:    gs.convertUser(reqCtx.User),
		HTTPClient:  gs.httpClient,
		Logger:      gs.logger,
	}

	result, err := gs.strategyManager.ExecuteStrategy(ctx, routeConfig.Strategy, strategyParams)
	if err != nil {
		gs.logger.Error("GraphQL strategy execution failed", err, map[string]interface{}{
			"request_id": reqCtx.RequestID,
			"strategy":   routeConfig.Strategy,
		})
		return &domain.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       map[string]string{"error": "GraphQL execution failed"},
		}, nil
	}

	return &domain.Response{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: result,
	}, nil
}

// convertUser converts domain user to ports user info
func (gs *GatewayService) convertUser(user *domain.User) *ports.UserInfo {
	if user == nil {
		return nil
	}
	return &ports.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
		Metadata: user.Metadata,
	}
}

// convertHTTPResponse converts http.Response to domain.Response
func (gs *GatewayService) convertHTTPResponse(httpResp *http.Response) (*domain.Response, error) {
	defer httpResp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert headers
	headers := make(map[string]string)
	for key, values := range httpResp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Detect binary content (e.g., images) and preserve raw bytes
	contentType := httpResp.Header.Get("Content-Type")
	isBinary := strings.HasPrefix(strings.ToLower(contentType), "image/") ||
		strings.HasPrefix(strings.ToLower(contentType), "application/octet-stream")

	if isBinary {
		return &domain.Response{
			StatusCode: httpResp.StatusCode,
			Headers:    headers,
			Body:       bodyBytes,
			Metadata:   map[string]interface{}{"is_binary": true},
		}, nil
	}

	// For non-binary content, try to parse as JSON; fallback to string
	var body interface{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			body = string(bodyBytes)
		}
	} else {
		body = map[string]interface{}{}
	}

	return &domain.Response{
		StatusCode: httpResp.StatusCode,
		Headers:    headers,
		Body:       body,
	}, nil
}

// createHTTPRequestFromContext creates an http.Request from RequestContext
func (gs *GatewayService) createHTTPRequestFromContext(reqCtx *domain.RequestContext) *http.Request {
	// Build URL
	requestURL := &url.URL{
		Scheme: "http",      // Default scheme
		Host:   "localhost", // Default host
		Path:   reqCtx.Path,
	}

	// Add query parameters
	if len(reqCtx.Query) > 0 {
		values := url.Values{}
		for key, value := range reqCtx.Query {
			values.Add(key, value)
		}
		requestURL.RawQuery = values.Encode()
	}

	// Create request body
	var body io.Reader
	if reqCtx.Body != nil {
		// Check if this is raw body data (e.g., multipart/form-data)
		if rawBody, ok := reqCtx.Body.([]byte); ok {
			body = bytes.NewReader(rawBody)
		} else {
			// Try to marshal as JSON for other cases
			if jsonBytes, err := json.Marshal(reqCtx.Body); err == nil {
				body = bytes.NewReader(jsonBytes)
			}
		}
	}

	// Create the HTTP request
	req, err := http.NewRequest(reqCtx.Method, requestURL.String(), body)
	if err != nil {
		// If creation fails, return a minimal request
		req, _ = http.NewRequest("GET", "/", nil)
	}

	// Add headers
	for key, value := range reqCtx.Headers {
		// Skip the custom header we added for raw body detection
		if key != "x-raw-body" {
			req.Header.Set(key, value)
		}
	}

	// Set content type if body exists and no content-type is set
	if body != nil && req.Header.Get("Content-Type") == "" {
		// Check if this is raw body data (multipart/form-data)
		if _, ok := reqCtx.Body.([]byte); ok {
			// For raw body, we should have received the content-type from the original request
			// If not, this is an error condition
			req.Header.Set("Content-Type", "application/octet-stream")
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	return req
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
