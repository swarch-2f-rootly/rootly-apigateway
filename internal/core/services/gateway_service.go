package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	// Check for API key first
	if apiKey, exists := reqCtx.Headers["x-api-key"]; exists {
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

	return nil, fmt.Errorf("no valid authentication provided")
}

// handleProxyMode handles proxy mode requests
func (gs *GatewayService) handleProxyMode(ctx context.Context, reqCtx *domain.RequestContext, routeConfig ports.RouteConfig) (*domain.Response, error) {
	serviceInfo, found := gs.configProvider.GetServiceConfig(routeConfig.Upstream)
	if !found {
		return &domain.Response{
			StatusCode: http.StatusBadGateway,
			Body:       map[string]string{"error": "Upstream service not configured"},
		}, nil
	}

	// Execute proxy strategy
	strategyParams := ports.StrategyParams{
		Request:     reqCtx.Request,
		RouteConfig: routeConfig,
		Services: map[string]ports.ServiceInfo{
			routeConfig.Upstream: *serviceInfo,
		},
		UserInfo:   gs.convertUser(reqCtx.User),
		HTTPClient: gs.httpClient,
		Logger:     gs.logger,
	}

	result, err := gs.strategyManager.ExecuteStrategy(ctx, "proxy", strategyParams)
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

	if httpResp, ok := result.(*http.Response); ok {
		return gs.convertHTTPResponse(httpResp)
	}

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

	// Execute logic strategy
	strategyParams := ports.StrategyParams{
		Request:     reqCtx.Request,
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

	// Execute GraphQL strategy
	strategyParams := ports.StrategyParams{
		Request:     reqCtx.Request,
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
	// Read and convert upstream response into a JSON-friendly structure
	// Copy headers we care about (e.g., Content-Type)
	headers := make(map[string]string)
	if ct := httpResp.Header.Get("Content-Type"); ct != "" {
		headers["Content-Type"] = ct
	}

	// Read full body
	defer httpResp.Body.Close()
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return &domain.Response{
			StatusCode: httpResp.StatusCode,
			Headers:    headers,
			Body:       map[string]string{"error": "failed to read upstream response"},
		}, nil
	}

	var body interface{}
	if len(bodyBytes) == 0 {
		body = map[string]interface{}{}
	} else {
		// Try to parse as JSON; if it fails, return raw string
		var jsonBody interface{}
		if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
			body = jsonBody
		} else {
			body = string(bodyBytes)
		}
	}

	return &domain.Response{
		StatusCode: httpResp.StatusCode,
		Headers:    headers,
		Body:       body,
	}, nil
}
