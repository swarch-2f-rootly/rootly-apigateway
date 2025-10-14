package strategies

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
)

// serviceCall represents a service call configuration for orchestration
type serviceCall struct {
	service  string
	endpoint string
	method   string
}

// ProxyStrategy implements simple reverse proxy functionality
type ProxyStrategy struct {
	name string
}

// NewProxyStrategy creates a new proxy strategy
func NewProxyStrategy() *ProxyStrategy {
	return &ProxyStrategy{
		name: "proxy",
	}
}

// GetName returns the strategy name
func (ps *ProxyStrategy) GetName() string {
	return ps.name
}

// Execute executes the proxy strategy
func (ps *ProxyStrategy) Execute(ctx context.Context, params ports.StrategyParams) (interface{}, error) {
	routeConfig := params.RouteConfig
	
	// Get target service info
	serviceInfo, exists := params.Services[routeConfig.Upstream]
	if !exists {
		return nil, fmt.Errorf("upstream service not found: %s", routeConfig.Upstream)
	}

	// Build target URL
	targetPath := routeConfig.TargetPath
	if targetPath == "" {
		targetPath = params.Request.URL.Path
	}

	// Replace path parameters
	targetPath = ps.replacePathParameters(targetPath, params.Request.URL.Path, routeConfig.Path)
	
	targetURL := serviceInfo.URL + targetPath
	if params.Request.URL.RawQuery != "" {
		targetURL += "?" + params.Request.URL.RawQuery
	}

	params.Logger.Debug("Proxying request", map[string]interface{}{
		"original_url": params.Request.URL.String(),
		"target_url":   targetURL,
		"method":       params.Request.Method,
	})

	// Create new request
	var body io.Reader
	if params.Request.Body != nil {
		bodyBytes, err := io.ReadAll(params.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, params.Request.Method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}

	// Copy headers (excluding host and hop-by-hop headers)
	for name, values := range params.Request.Header {
		if ps.shouldForwardHeader(name) {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}
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
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}

	params.Logger.Debug("Proxy request completed", map[string]interface{}{
		"status_code": resp.StatusCode,
		"target_url":  targetURL,
	})

	return resp, nil
}

// replacePathParameters replaces path parameters in target path
func (ps *ProxyStrategy) replacePathParameters(targetPath, requestPath, routePath string) string {
	routeParts := strings.Split(strings.Trim(routePath, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")

	if len(routeParts) != len(requestParts) {
		return targetPath
	}

	result := targetPath
	for i, routePart := range routeParts {
		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			paramName := routePart
			paramValue := requestParts[i]
			result = strings.ReplaceAll(result, paramName, paramValue)
		}
	}

	return result
}

// shouldForwardHeader determines if a header should be forwarded
func (ps *ProxyStrategy) shouldForwardHeader(name string) bool {
	// Don't forward hop-by-hop headers
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	lowerName := strings.ToLower(name)
	for _, header := range hopByHopHeaders {
		if lowerName == strings.ToLower(header) {
			return false
		}
	}

	return true
}

// DashboardOrchestratorStrategy orchestrates multiple service calls for dashboard data
type DashboardOrchestratorStrategy struct {
	name string
}

// NewDashboardOrchestratorStrategy creates a new dashboard orchestrator strategy
func NewDashboardOrchestratorStrategy() *DashboardOrchestratorStrategy {
	return &DashboardOrchestratorStrategy{
		name: "dashboard_orchestrator",
	}
}

// GetName returns the strategy name
func (dos *DashboardOrchestratorStrategy) GetName() string {
	return dos.name
}

// Execute executes the dashboard orchestrator strategy
func (dos *DashboardOrchestratorStrategy) Execute(ctx context.Context, params ports.StrategyParams) (interface{}, error) {
	params.Logger.Info("Executing dashboard orchestrator", map[string]interface{}{
		"user_id": func() string {
			if params.UserInfo != nil {
				return params.UserInfo.ID
			}
			return "anonymous"
		}(),
	})

	// Prepare parallel service calls
	results := make(map[string]interface{})
	errors := make(map[string]error)

	// Create channels for parallel execution
	type serviceResult struct {
		service string
		data    interface{}
		err     error
	}

	resultChan := make(chan serviceResult, len(params.RouteConfig.Upstreams))

	// Execute calls in parallel
	for _, upstream := range params.RouteConfig.Upstreams {
		go func(up ports.UpstreamConfig) {
			data, err := dos.callService(ctx, up, params)
			resultChan <- serviceResult{
				service: up.Service,
				data:    data,
				err:     err,
			}
		}(upstream)
	}

	// Collect results
	for i := 0; i < len(params.RouteConfig.Upstreams); i++ {
		result := <-resultChan
		if result.err != nil {
			errors[result.service] = result.err
			params.Logger.Error("Service call failed", result.err, map[string]interface{}{
				"service": result.service,
			})
		} else {
			results[result.service] = result.data
		}
	}

	// Build dashboard response
	dashboardData := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      results,
	}

	if len(errors) > 0 {
		dashboardData["errors"] = errors
		params.Logger.Warn("Dashboard data partially available", map[string]interface{}{
			"successful_services": len(results),
			"failed_services":     len(errors),
		})
	}

	return dashboardData, nil
}

// callService makes a call to a specific service
func (dos *DashboardOrchestratorStrategy) callService(ctx context.Context, upstream ports.UpstreamConfig, params ports.StrategyParams) (interface{}, error) {
	serviceInfo, exists := params.Services[upstream.Service]
	if !exists {
		return nil, fmt.Errorf("service not configured: %s", upstream.Service)
	}

	targetURL := serviceInfo.URL + upstream.Endpoint
	method := upstream.Method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if user is authenticated
	if params.UserInfo != nil {
		// Add user context headers
		req.Header.Set("X-User-ID", params.UserInfo.ID)
		req.Header.Set("X-User-Email", params.UserInfo.Email)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("service returned error status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// If JSON parsing fails, return raw string
		return string(body), nil
	}

	return result, nil
}

// PlantFullReportStrategy orchestrates calls for a complete plant report
type PlantFullReportStrategy struct {
	name string
}

// NewPlantFullReportStrategy creates a new plant full report strategy
func NewPlantFullReportStrategy() *PlantFullReportStrategy {
	return &PlantFullReportStrategy{
		name: "plant_full_report",
	}
}

// GetName returns the strategy name
func (pfrs *PlantFullReportStrategy) GetName() string {
	return pfrs.name
}

// Execute executes the plant full report strategy
func (pfrs *PlantFullReportStrategy) Execute(ctx context.Context, params ports.StrategyParams) (interface{}, error) {
	// Extract plant ID from request path
	plantID := pfrs.extractPlantID(params.Request.URL.Path)
	if plantID == "" {
		return nil, fmt.Errorf("plant ID not found in path")
	}

	params.Logger.Info("Generating plant full report", map[string]interface{}{
		"plant_id": plantID,
		"user_id": func() string {
			if params.UserInfo != nil {
				return params.UserInfo.ID
			}
			return "anonymous"
		}(),
	})

	// Prepare service calls with plant ID
	calls := []serviceCall{}
	for _, upstream := range params.RouteConfig.Upstreams {
		endpoint := strings.ReplaceAll(upstream.Endpoint, "{id}", plantID)
		calls = append(calls, serviceCall{
			service:  upstream.Service,
			endpoint: endpoint,
			method:   upstream.Method,
		})
	}

	// Execute calls in parallel
	results := make(map[string]interface{})
	errors := make(map[string]error)

	type serviceResult struct {
		service string
		data    interface{}
		err     error
	}

	resultChan := make(chan serviceResult, len(calls))

	for _, call := range calls {
		go func(c serviceCall) {
			data, err := pfrs.callServiceForPlant(ctx, c, plantID, params)
			resultChan <- serviceResult{
				service: c.service,
				data:    data,
				err:     err,
			}
		}(call)
	}

	// Collect results
	for i := 0; i < len(calls); i++ {
		result := <-resultChan
		if result.err != nil {
			errors[result.service] = result.err
			params.Logger.Error("Plant service call failed", result.err, map[string]interface{}{
				"service":  result.service,
				"plant_id": plantID,
			})
		} else {
			results[result.service] = result.data
		}
	}

	// Build comprehensive plant report
	report := map[string]interface{}{
		"plant_id":  plantID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"report": map[string]interface{}{
			"plant_info": results["plant_management"],
			"analytics":  results["analytics"],
			"measurements": results["data_management"],
		},
	}

	if len(errors) > 0 {
		report["errors"] = errors
		// Check if critical data is missing
		if _, hasPlantInfo := results["plant_management"]; !hasPlantInfo {
			return nil, fmt.Errorf("failed to retrieve critical plant information")
		}
	}

	return report, nil
}

// extractPlantID extracts plant ID from the request path
func (pfrs *PlantFullReportStrategy) extractPlantID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "plant" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// callServiceForPlant makes a service call for plant-specific data
func (pfrs *PlantFullReportStrategy) callServiceForPlant(ctx context.Context, call serviceCall, plantID string, params ports.StrategyParams) (interface{}, error) {
	serviceInfo, exists := params.Services[call.service]
	if !exists {
		return nil, fmt.Errorf("service not configured: %s", call.service)
	}

	targetURL := serviceInfo.URL + call.endpoint
	method := call.method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and plant context headers
	if params.UserInfo != nil {
		req.Header.Set("X-User-ID", params.UserInfo.ID)
		req.Header.Set("X-User-Email", params.UserInfo.Email)
	}
	req.Header.Set("X-Plant-ID", plantID)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("service returned error status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return string(body), nil
	}

	return result, nil
}