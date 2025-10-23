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

	params.Logger.Info("🌐 Proxying request", map[string]interface{}{
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
		params.Logger.Error("❌ Proxy request failed", err, map[string]interface{}{
			"target_url": targetURL,
			"method":     params.Request.Method,
		})
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}

	// Log detailed response information
	params.Logger.Info("📥 Proxy response received", map[string]interface{}{
		"status_code":    resp.StatusCode,
		"target_url":     targetURL,
		"content_length": resp.ContentLength,
		"content_type":   resp.Header.Get("Content-Type"),
		"server":         resp.Header.Get("Server"),
	})

	// Log response body size without reading it (to avoid consuming the stream)
	if resp.ContentLength > 0 {
		params.Logger.Debug("📊 Response size details", map[string]interface{}{
			"content_length": resp.ContentLength,
			"target_url":     targetURL,
		})
	} else if resp.ContentLength == -1 {
		params.Logger.Debug("📊 Response with unknown size", map[string]interface{}{
			"target_url":        targetURL,
			"transfer_encoding": resp.Header.Get("Transfer-Encoding"),
		})
	}

	return resp, nil
}

// replacePathParameters replaces path parameters in target path
func (ps *ProxyStrategy) replacePathParameters(targetPath, requestPath, routePath string) string {
	routeParts := strings.Split(strings.Trim(routePath, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")

	// Handle wildcard routes (ending with *)
	if len(routeParts) > 0 && routeParts[len(routeParts)-1] == "*" {
		routePrefix := routeParts[:len(routeParts)-1]
		if len(requestParts) >= len(routePrefix) {
			// Replace parameters in the prefix
			result := targetPath
			for i, routePart := range routePrefix {
				if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
					paramValue := requestParts[i]
					result = strings.ReplaceAll(result, routePart, paramValue)
				}
			}

			// Handle wildcard (*) in target path
			if strings.Contains(result, "*") {
				// Extract the remaining path after the prefix
				remainingParts := requestParts[len(routePrefix):]
				remainingPath := strings.Join(remainingParts, "/")
				result = strings.ReplaceAll(result, "*", remainingPath)
			}

			return result
		}
		return targetPath
	}

	// Handle exact match routes
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

// UserProfileOrchestratorStrategy orchestrates calls for complete user profile
type UserProfileOrchestratorStrategy struct {
	name string
}

// NewUserProfileOrchestratorStrategy creates a new user profile orchestrator strategy
func NewUserProfileOrchestratorStrategy() *UserProfileOrchestratorStrategy {
	return &UserProfileOrchestratorStrategy{
		name: "user_profile_orchestrator",
	}
}

// GetName returns the strategy name
func (upos *UserProfileOrchestratorStrategy) GetName() string {
	return upos.name
}

// Execute executes the user profile orchestrator strategy
func (upos *UserProfileOrchestratorStrategy) Execute(ctx context.Context, params ports.StrategyParams) (interface{}, error) {
	// Extract user ID from authenticated user or from path
	userID := upos.extractUserID(params)
	if userID == "" {
		return nil, fmt.Errorf("user ID not found")
	}

	params.Logger.Info("🔍 Fetching user profile", map[string]interface{}{
		"user_id": userID,
	})

	// Define service calls for user profile data
	type serviceResult struct {
		key  string
		data interface{}
		err  error
	}

	resultChan := make(chan serviceResult, 3)

	// 1. Fetch user basic information from auth service
	go func() {
		data, err := upos.fetchUserInfo(ctx, userID, params)
		resultChan <- serviceResult{key: "user_info", data: data, err: err}
	}()

	// 2. Fetch user's plants from plant management service
	go func() {
		data, err := upos.fetchUserPlants(ctx, userID, params)
		resultChan <- serviceResult{key: "plants", data: data, err: err}
	}()

	// 3. Fetch user's devices from plant management service
	go func() {
		data, err := upos.fetchUserDevices(ctx, userID, params)
		resultChan <- serviceResult{key: "devices", data: data, err: err}
	}()

	// Collect results
	results := make(map[string]interface{})
	errors := make(map[string]string)

	for i := 0; i < 3; i++ {
		result := <-resultChan
		if result.err != nil {
			errors[result.key] = result.err.Error()
			params.Logger.Warn(fmt.Sprintf("Failed to fetch %s", result.key), map[string]interface{}{
				"user_id": userID,
				"error":   result.err.Error(),
			})
		} else {
			results[result.key] = result.data
		}
	}

	// Check if critical data is missing (user_info is required)
	if _, hasUserInfo := results["user_info"]; !hasUserInfo {
		return nil, fmt.Errorf("failed to retrieve user information")
	}

	// Build comprehensive profile response
	profile := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"user":      results["user_info"],
		"plants":    results["plants"],
		"devices":   results["devices"],
		"stats": map[string]interface{}{
			"total_plants":  upos.countItems(results["plants"]),
			"total_devices": upos.countItems(results["devices"]),
		},
	}

	if len(errors) > 0 {
		profile["partial_errors"] = errors
		params.Logger.Info("✅ Profile loaded with some partial errors", map[string]interface{}{
			"user_id":      userID,
			"errors_count": len(errors),
		})
	} else {
		params.Logger.Info("✅ Profile loaded successfully", map[string]interface{}{
			"user_id": userID,
		})
	}

	return profile, nil
}

// extractUserID extracts user ID from authenticated user or request path
func (upos *UserProfileOrchestratorStrategy) extractUserID(params ports.StrategyParams) string {
	// Priority 1: From authenticated user context
	if params.UserInfo != nil && params.UserInfo.ID != "" {
		return params.UserInfo.ID
	}

	// Priority 2: From URL path parameter (e.g., /api/v1/profile/{user_id})
	parts := strings.Split(strings.Trim(params.Request.URL.Path, "/"), "/")
	for i, part := range parts {
		if part == "profile" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

// fetchUserInfo retrieves user information from auth service
func (upos *UserProfileOrchestratorStrategy) fetchUserInfo(ctx context.Context, userID string, params ports.StrategyParams) (interface{}, error) {
	serviceInfo, exists := params.Services["auth"]
	if !exists {
		return nil, fmt.Errorf("auth service not configured")
	}

	targetURL := fmt.Sprintf("%s/api/v1/users/%s", serviceInfo.URL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Forward authorization header
	if authHeader := params.Request.Header.Get("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// fetchUserPlants retrieves user's plants from plant management service
func (upos *UserProfileOrchestratorStrategy) fetchUserPlants(ctx context.Context, userID string, params ports.StrategyParams) (interface{}, error) {
	serviceInfo, exists := params.Services["plant_management"]
	if !exists {
		return nil, fmt.Errorf("plant_management service not configured")
	}

	targetURL := fmt.Sprintf("%s/api/v1/plants/users/%s", serviceInfo.URL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Forward authorization header
	if authHeader := params.Request.Header.Get("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// If not found, return empty array instead of error
	if resp.StatusCode == 404 {
		return []interface{}{}, nil
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("plant service returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// fetchUserDevices retrieves user's devices from plant management service
func (upos *UserProfileOrchestratorStrategy) fetchUserDevices(ctx context.Context, userID string, params ports.StrategyParams) (interface{}, error) {
	serviceInfo, exists := params.Services["plant_management"]
	if !exists {
		return nil, fmt.Errorf("plant_management service not configured")
	}

	targetURL := fmt.Sprintf("%s/api/v1/devices/users/%s", serviceInfo.URL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Forward authorization header
	if authHeader := params.Request.Header.Get("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// If not found, return empty array instead of error
	if resp.StatusCode == 404 {
		return []interface{}{}, nil
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device service returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// countItems counts the number of items in a result (handles both arrays and maps)
func (upos *UserProfileOrchestratorStrategy) countItems(data interface{}) int {
	if data == nil {
		return 0
	}

	switch v := data.(type) {
	case []interface{}:
		return len(v)
	case map[string]interface{}:
		if items, ok := v["items"].([]interface{}); ok {
			return len(items)
		}
		return 1
	default:
		return 0
	}
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
			"plant_info":   results["plant_management"],
			"analytics":    results["analytics"],
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
