package strategies

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
)

// LocalSchemaStrategy handles GraphQL requests with local schema resolution
type LocalSchemaStrategy struct {
	name string
}

// NewLocalSchemaStrategy creates a new local schema strategy
func NewLocalSchemaStrategy() *LocalSchemaStrategy {
	return &LocalSchemaStrategy{
		name: "local_schema",
	}
}

// GetName returns the strategy name
func (lss *LocalSchemaStrategy) GetName() string {
	return lss.name
}

// Execute executes the local schema strategy
func (lss *LocalSchemaStrategy) Execute(ctx context.Context, params ports.StrategyParams) (interface{}, error) {
	// Parse GraphQL request
	var gqlRequest GraphQLRequest
	if params.Request.Body != nil {
		body, err := io.ReadAll(params.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read GraphQL request body: %w", err)
		}

		if err := json.Unmarshal(body, &gqlRequest); err != nil {
			return nil, fmt.Errorf("failed to parse GraphQL request: %w", err)
		}
	}

	params.Logger.Info("Processing GraphQL query", map[string]interface{}{
		"operation_name": gqlRequest.OperationName,
		"has_variables":  len(gqlRequest.Variables) > 0,
		"query_length":   len(gqlRequest.Query),
	})

	// Route based on operation type
	response, err := lss.routeGraphQLOperation(ctx, gqlRequest, params)
	if err != nil {
		return lss.buildErrorResponse(err), nil
	}

	return response, nil
}

// routeGraphQLOperation routes GraphQL operations to appropriate services
func (lss *LocalSchemaStrategy) routeGraphQLOperation(ctx context.Context, request GraphQLRequest, params ports.StrategyParams) (interface{}, error) {
	// Simple operation routing based on query content
	// In a real implementation, you would use a proper GraphQL parser
	query := request.Query

	switch {
	case lss.containsField(query, "analytics", "metrics", "measurements"):
		return lss.callAnalyticsService(ctx, request, params)
	case lss.containsField(query, "plants", "devices"):
		return lss.callPlantManagementService(ctx, request, params)
	case lss.containsField(query, "users", "auth"):
		return lss.callAuthService(ctx, request, params)
	case lss.containsField(query, "dashboard"):
		return lss.orchestrateDashboardQuery(ctx, request, params)
	default:
		return lss.handleIntrospectionOrDefault(ctx, request, params)
	}
}

// containsField checks if the query contains specific fields
func (lss *LocalSchemaStrategy) containsField(query string, fields ...string) bool {
	for _, field := range fields {
		if contains(query, field) {
			return true
		}
	}
	return false
}

// callAnalyticsService forwards GraphQL query to analytics service
func (lss *LocalSchemaStrategy) callAnalyticsService(ctx context.Context, request GraphQLRequest, params ports.StrategyParams) (interface{}, error) {
	serviceInfo, exists := params.Services["analytics"]
	if !exists {
		return nil, fmt.Errorf("analytics service not configured")
	}

	return lss.forwardGraphQLRequest(ctx, request, serviceInfo, "/graphql", params)
}

// callPlantManagementService forwards GraphQL query to plant management service
func (lss *LocalSchemaStrategy) callPlantManagementService(ctx context.Context, request GraphQLRequest, params ports.StrategyParams) (interface{}, error) {
	serviceInfo, exists := params.Services["plant_management"]
	if !exists {
		return nil, fmt.Errorf("plant management service not configured")
	}

	return lss.forwardGraphQLRequest(ctx, request, serviceInfo, "/graphql", params)
}

// callAuthService forwards GraphQL query to auth service
func (lss *LocalSchemaStrategy) callAuthService(ctx context.Context, request GraphQLRequest, params ports.StrategyParams) (interface{}, error) {
	serviceInfo, exists := params.Services["auth"]
	if !exists {
		return nil, fmt.Errorf("auth service not configured")
	}

	return lss.forwardGraphQLRequest(ctx, request, serviceInfo, "/graphql", params)
}

// orchestrateDashboardQuery handles dashboard queries that require multiple services
func (lss *LocalSchemaStrategy) orchestrateDashboardQuery(ctx context.Context, request GraphQLRequest, params ports.StrategyParams) (interface{}, error) {
	// For dashboard queries, we need to orchestrate multiple service calls
	// This is a simplified example - real implementation would parse the GraphQL query properly
	
	results := make(map[string]interface{})
	
	// Call analytics for metrics
	if analyticsService, exists := params.Services["analytics"]; exists {
		analyticsQuery := GraphQLRequest{
			Query: `query { metrics { temperature humidity lightLevel } }`,
		}
		analyticsResult, err := lss.forwardGraphQLRequest(ctx, analyticsQuery, analyticsService, "/graphql", params)
		if err == nil {
			results["analytics"] = analyticsResult
		}
	}

	// Call plant management for plants
	if plantService, exists := params.Services["plant_management"]; exists {
		plantsQuery := GraphQLRequest{
			Query: `query { plants { id name type status } }`,
		}
		plantsResult, err := lss.forwardGraphQLRequest(ctx, plantsQuery, plantService, "/graphql", params)
		if err == nil {
			results["plants"] = plantsResult
		}
	}

	return map[string]interface{}{
		"data": map[string]interface{}{
			"dashboard": results,
		},
	}, nil
}

// handleIntrospectionOrDefault handles introspection queries or returns schema
func (lss *LocalSchemaStrategy) handleIntrospectionOrDefault(ctx context.Context, request GraphQLRequest, params ports.StrategyParams) (interface{}, error) {
	if contains(request.Query, "__schema") || contains(request.Query, "__type") {
		// Return a basic schema for introspection
		return lss.buildIntrospectionResponse(), nil
	}

	// Default error for unknown queries
	return nil, fmt.Errorf("unknown GraphQL operation")
}

// forwardGraphQLRequest forwards a GraphQL request to a specific service
func (lss *LocalSchemaStrategy) forwardGraphQLRequest(ctx context.Context, request GraphQLRequest, serviceInfo ports.ServiceInfo, endpoint string, params ports.StrategyParams) (interface{}, error) {
	targetURL := serviceInfo.URL + endpoint

	// Serialize request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize GraphQL request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers
	if params.UserInfo != nil {
		req.Header.Set("X-User-ID", params.UserInfo.ID)
		req.Header.Set("X-User-Email", params.UserInfo.Email)
	}

	// Set timeout
	timeout := 30 * time.Second
	if serviceInfo.Timeout != "" {
		if parsedTimeout, err := time.ParseDuration(serviceInfo.Timeout); err == nil {
			timeout = parsedTimeout
		}
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GraphQL response: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return result, nil
}

// buildErrorResponse builds a GraphQL error response
func (lss *LocalSchemaStrategy) buildErrorResponse(err error) interface{} {
	return map[string]interface{}{
		"errors": []map[string]interface{}{
			{
				"message": err.Error(),
			},
		},
	}
}

// buildIntrospectionResponse builds a basic introspection response
func (lss *LocalSchemaStrategy) buildIntrospectionResponse() interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"__schema": map[string]interface{}{
				"types": []map[string]interface{}{
					{
						"name": "Query",
						"kind": "OBJECT",
						"fields": []map[string]interface{}{
							{"name": "analytics", "type": map[string]interface{}{"name": "Analytics"}},
							{"name": "plants", "type": map[string]interface{}{"name": "[Plant]"}},
							{"name": "dashboard", "type": map[string]interface{}{"name": "Dashboard"}},
						},
					},
				},
			},
		},
	}
}

// GraphQLProxyStrategy handles GraphQL requests by proxying to upstream services
type GraphQLProxyStrategy struct {
	name string
}

// NewGraphQLProxyStrategy creates a new GraphQL proxy strategy
func NewGraphQLProxyStrategy() *GraphQLProxyStrategy {
	return &GraphQLProxyStrategy{
		name: "graphql_proxy",
	}
}

// GetName returns the strategy name
func (gps *GraphQLProxyStrategy) GetName() string {
	return gps.name
}

// Execute executes the GraphQL proxy strategy
func (gps *GraphQLProxyStrategy) Execute(ctx context.Context, params ports.StrategyParams) (interface{}, error) {
	routeConfig := params.RouteConfig
	
	// Get target service info
	serviceInfo, exists := params.Services[routeConfig.Upstream]
	if !exists {
		return nil, fmt.Errorf("upstream service not found: %s", routeConfig.Upstream)
	}

	// Parse GraphQL request
	var gqlRequest GraphQLRequest
	if params.Request.Body != nil {
		body, err := io.ReadAll(params.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read GraphQL request body: %w", err)
		}

		if err := json.Unmarshal(body, &gqlRequest); err != nil {
			return nil, fmt.Errorf("failed to parse GraphQL request: %w", err)
		}
	}

	params.Logger.Info("Proxying GraphQL request", map[string]interface{}{
		"upstream":       routeConfig.Upstream,
		"operation_name": gqlRequest.OperationName,
	})

	// Forward to upstream service
	targetURL := serviceInfo.URL + routeConfig.TargetPath
	if routeConfig.TargetPath == "" {
		targetURL = serviceInfo.URL + "/graphql"
	}

	return gps.forwardRequest(ctx, gqlRequest, targetURL, params)
}

// forwardRequest forwards the GraphQL request to upstream
func (gps *GraphQLProxyStrategy) forwardRequest(ctx context.Context, request GraphQLRequest, targetURL string, params ports.StrategyParams) (interface{}, error) {
	// Serialize request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize GraphQL request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Copy relevant headers from original request
	if params.Request.Header.Get("Authorization") != "" {
		req.Header.Set("Authorization", params.Request.Header.Get("Authorization"))
	}

	// Add user context headers
	if params.UserInfo != nil {
		req.Header.Set("X-User-ID", params.UserInfo.ID)
		req.Header.Set("X-User-Email", params.UserInfo.Email)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GraphQL proxy request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GraphQL response: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return result, nil
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || 
		    (len(str) > len(substr) && 
		     (str[:len(substr)] == substr || 
		      str[len(str)-len(substr):] == substr || 
		      containsHelper(str, substr))))
}

func containsHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}