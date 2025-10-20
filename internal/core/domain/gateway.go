package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Gateway represents the main API Gateway entity
type Gateway struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
	Routes      []Route   `json:"routes"`
	Services    []Service `json:"services"`
}

// Route represents a configured route in the gateway
type Route struct {
	ID           string                 `json:"id"`
	Path         string                 `json:"path"`
	Method       string                 `json:"method"`
	Mode         RouteMode              `json:"mode"`
	Strategy     string                 `json:"strategy,omitempty"`
	Upstream     string                 `json:"upstream,omitempty"`
	TargetPath   string                 `json:"target_path,omitempty"`
	AuthRequired bool                   `json:"auth_required"`
	Upstreams    []Upstream             `json:"upstreams,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// RouteMode represents the different modes a route can operate in
type RouteMode string

const (
	// ProxyMode routes requests directly to an upstream service
	ProxyMode RouteMode = "proxy"
	// LogicMode applies business logic and may orchestrate multiple services
	LogicMode RouteMode = "logic"
	// GraphQLMode handles GraphQL requests
	GraphQLMode RouteMode = "graphql"
)

// Upstream represents an upstream service configuration
type Upstream struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
	Method   string `json:"method,omitempty"`
}

// Service represents a backend service
type Service struct {
	Name        string        `json:"name"`
	URL         string        `json:"url"`
	Status      ServiceStatus `json:"status"`
	Timeout     time.Duration `json:"timeout"`
	HealthCheck string        `json:"health_check,omitempty"`
	LastChecked time.Time     `json:"last_checked,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ServiceStatus represents the status of a backend service
type ServiceStatus string

const (
	ServiceStatusHealthy   ServiceStatus = "healthy"
	ServiceStatusUnhealthy ServiceStatus = "unhealthy"
	ServiceStatusUnknown   ServiceStatus = "unknown"
)

// RequestContext represents the context of an incoming request
type RequestContext struct {
	RequestID   string                 `json:"request_id"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Headers     map[string]string      `json:"headers"`
	Query       map[string]string      `json:"query"`
	Body        interface{}            `json:"body,omitempty"`
	User        *User                  `json:"user,omitempty"`
	Route       *Route                 `json:"route,omitempty"`
	StartTime   time.Time              `json:"start_time"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// User represents an authenticated user
type User struct {
	ID       string                 `json:"id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Roles    []string               `json:"roles"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Response represents a gateway response
type Response struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Strategy represents a routing strategy
type Strategy struct {
	Name              string                 `json:"name"`
	Type              StrategyType           `json:"type"`
	Config            map[string]interface{} `json:"config"`
	Description       string                 `json:"description,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// StrategyType represents the type of strategy
type StrategyType string

const (
	ProxyStrategy              StrategyType = "proxy"
	DashboardOrchestrator      StrategyType = "dashboard_orchestrator"
	PlantFullReport            StrategyType = "plant_full_report"
	UserProfileOrchestrator    StrategyType = "user_profile_orchestrator"
	LocalSchema                StrategyType = "local_schema"
)

// Validate validates the route configuration
func (r *Route) Validate() error {
	if r.Path == "" {
		return errors.New("route path cannot be empty")
	}
	
	if r.Method == "" {
		return errors.New("route method cannot be empty")
	}
	
	if r.Mode == "" {
		return errors.New("route mode cannot be empty")
	}
	
	switch r.Mode {
	case ProxyMode:
		if r.Upstream == "" {
			return errors.New("proxy mode requires upstream configuration")
		}
	case LogicMode:
		if r.Strategy == "" {
			return errors.New("logic mode requires strategy configuration")
		}
		if len(r.Upstreams) == 0 {
			return errors.New("logic mode requires at least one upstream")
		}
	case GraphQLMode:
		if r.Strategy == "" {
			return errors.New("graphql mode requires strategy configuration")
		}
	default:
		return fmt.Errorf("unsupported route mode: %s", r.Mode)
	}
	
	return nil
}

// IsHealthy checks if the service is healthy
func (s *Service) IsHealthy() bool {
	return s.Status == ServiceStatusHealthy
}

// NeedsHealthCheck determines if the service needs a health check
func (s *Service) NeedsHealthCheck(interval time.Duration) bool {
	return time.Since(s.LastChecked) > interval
}

// AddRole adds a role to the user if not already present
func (u *User) AddRole(role string) {
	for _, r := range u.Roles {
		if r == role {
			return
		}
	}
	u.Roles = append(u.Roles, role)
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (u *User) HasAnyRole(roles []string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// MatchesPath checks if a request path matches the route path pattern
func (r *Route) MatchesPath(requestPath string) bool {
	// Simple path matching with support for path parameters
	routeParts := strings.Split(strings.Trim(r.Path, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	
	if len(routeParts) != len(requestParts) {
		return false
	}
	
	for i, routePart := range routeParts {
		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			// This is a path parameter, skip validation
			continue
		}
		if routePart != requestParts[i] {
			return false
		}
	}
	
	return true
}

// ExtractPathParams extracts path parameters from a request path
func (r *Route) ExtractPathParams(requestPath string) map[string]string {
	params := make(map[string]string)
	
	routeParts := strings.Split(strings.Trim(r.Path, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	
	if len(routeParts) != len(requestParts) {
		return params
	}
	
	for i, routePart := range routeParts {
		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			paramName := strings.Trim(routePart, "{}")
			params[paramName] = requestParts[i]
		}
	}
	
	return params
}

// BuildTargetURL builds the target URL for proxy requests
func (r *Route) BuildTargetURL(baseURL, requestPath string) string {
	if r.TargetPath == "" {
		return baseURL + requestPath
	}
	
	// Replace path parameters in target path
	targetPath := r.TargetPath
	params := r.ExtractPathParams(requestPath)
	
	for key, value := range params {
		targetPath = strings.ReplaceAll(targetPath, "{"+key+"}", value)
	}
	
	return baseURL + targetPath
}