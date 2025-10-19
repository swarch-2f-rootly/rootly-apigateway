package ports

import (
	"context"
	"net/http"
)

// HTTPClient defines the port for HTTP client operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// RouteHandler defines the port for handling different route modes
type RouteHandler interface {
	HandleProxy(ctx context.Context, req *http.Request, routeConfig RouteConfig) (*http.Response, error)
	HandleLogic(ctx context.Context, req *http.Request, routeConfig RouteConfig) (interface{}, error)
	HandleGraphQL(ctx context.Context, req *http.Request, routeConfig RouteConfig) (interface{}, error)
}

// RouteConfig represents the configuration for a specific route
type RouteConfig struct {
	Path         string
	Method       string
	Mode         string
	Strategy     string
	Upstream     string
	TargetPath   string
	AuthRequired bool
	Upstreams    []UpstreamConfig
	Metadata     map[string]interface{}
}

// UpstreamConfig represents configuration for upstream services
type UpstreamConfig struct {
	Service  string
	Endpoint string
	Method   string
}

// AuthService defines the port for authentication operations
type AuthService interface {
	ValidateAPIKey(ctx context.Context, apiKey string) (bool, error)
	ValidateJWT(ctx context.Context, token string) (*UserInfo, error)
	GenerateJWT(ctx context.Context, userInfo *UserInfo) (string, error)
}

// UserInfo represents authenticated user information
type UserInfo struct {
	ID       string                 `json:"id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Roles    []string               `json:"roles"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Logger defines the port for logging operations
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
}

// StrategyManager defines the port for managing routing strategies
type StrategyManager interface {
	RegisterStrategy(name string, strategy RouteStrategy)
	GetStrategy(name string) (RouteStrategy, bool)
	ExecuteStrategy(ctx context.Context, strategyName string, params StrategyParams) (interface{}, error)
}

// RouteStrategy defines the interface for different routing strategies
type RouteStrategy interface {
	Execute(ctx context.Context, params StrategyParams) (interface{}, error)
	GetName() string
}

// StrategyParams contains parameters for strategy execution
type StrategyParams struct {
	Request      *http.Request
	RouteConfig  RouteConfig
	Services     map[string]ServiceInfo
	UserInfo     *UserInfo
	HTTPClient   HTTPClient
	Logger       Logger
}

// ServiceInfo contains information about a backend service
type ServiceInfo struct {
	Name    string
	URL     string
	Timeout string
}

// ServiceOrchestrator defines the port for orchestrating multiple service calls
type ServiceOrchestrator interface {
	OrchestrateCalls(ctx context.Context, calls []ServiceCall) (map[string]interface{}, error)
}

// ServiceCall represents a call to a backend service
type ServiceCall struct {
	Service    string
	Endpoint   string
	Method     string
	Body       interface{}
	Headers    map[string]string
	Timeout    string
	Parallel   bool
}

// HealthChecker defines the port for health checking operations
type HealthChecker interface {
	CheckHealth(ctx context.Context, serviceName string) (HealthStatus, error)
	CheckAllServices(ctx context.Context) (map[string]HealthStatus, error)
}

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Status    string                 `json:"status"`    // "healthy", "unhealthy", "unknown"
	Message   string                 `json:"message,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MetricsCollector defines the port for collecting metrics
type MetricsCollector interface {
	IncrementCounter(name string, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
}

// ConfigProvider defines the port for configuration management
type ConfigProvider interface {
	GetRouteConfig(path string, method string) (*RouteConfig, bool)
	GetServiceConfig(serviceName string) (*ServiceInfo, bool)
	GetStrategyConfig(strategyName string) (map[string]interface{}, bool)
	ReloadConfig() error
}