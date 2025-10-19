package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowAllOrigins bool     `yaml:"allow_all_origins"`
	AllowedOrigins  []string `yaml:"allowed_origins"`
	AllowedMethods  []string `yaml:"allowed_methods"`
	AllowedHeaders  []string `yaml:"allowed_headers"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// ServiceConfig holds service endpoint configuration
type ServiceConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// RouteConfig represents a route configuration
type RouteConfig struct {
	Path         string                 `yaml:"path"`
	Method       string                 `yaml:"method"`
	Mode         string                 `yaml:"mode"` // proxy, logic, graphql
	Strategy     string                 `yaml:"strategy,omitempty"`
	Upstream     string                 `yaml:"upstream,omitempty"`
	TargetPath   string                 `yaml:"target_path,omitempty"`
	AuthRequired bool                   `yaml:"auth_required"`
	Upstreams    []UpstreamConfig       `yaml:"upstreams,omitempty"`
	Metadata     map[string]interface{} `yaml:"metadata,omitempty"`
}

// UpstreamConfig represents upstream service configuration for logic mode
type UpstreamConfig struct {
	Service  string `yaml:"service"`
	Endpoint string `yaml:"endpoint"`
	Method   string `yaml:"method,omitempty"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	APIKeyHeader  string        `yaml:"api_key_header"`
	JWTSecret     string        `yaml:"jwt_secret"`
	JWTExpiration time.Duration `yaml:"jwt_expiration"`
}

// StrategyConfig holds strategy-specific configuration
type StrategyConfig struct {
	Timeout          time.Duration `yaml:"timeout,omitempty"`
	ParallelRequests bool          `yaml:"parallel_requests,omitempty"`
	FailurePolicy    string        `yaml:"failure_policy,omitempty"`
	// GraphQL specific
	IntrospectionEnabled bool `yaml:"introspection_enabled,omitempty"`
	PlaygroundEnabled    bool `yaml:"playground_enabled,omitempty"`
	// Proxy specific
	PreserveHeaders bool          `yaml:"preserve_headers,omitempty"`
	ProxyTimeout    time.Duration `yaml:"proxy_timeout,omitempty"`
}

// Config holds all configuration for the API Gateway
type Config struct {
	Server     ServerConfig              `yaml:"server"`
	CORS       CORSConfig                `yaml:"cors"`
	Logging    LoggingConfig             `yaml:"logging"`
	Services   map[string]ServiceConfig  `yaml:"services"`
	Routes     []RouteConfig             `yaml:"routes"`
	Auth       AuthConfig                `yaml:"auth"`
	Strategies map[string]StrategyConfig `yaml:"strategies"`

	// Legacy fields for backward compatibility
	AnalyticsServiceURL         string
	AuthServiceURL              string
	DataManagementServiceURL    string
	PlantManagementServiceURL   string
	Port                        string
	GinMode                     string
	GraphQLPlaygroundEnabled    bool
	GraphQLIntrospectionEnabled bool
	CORSAllowAllOrigins         bool
	LogLevel                    string
	LogFormat                   string
}

// LoadConfig loads configuration from YAML file and environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config := &Config{}

	// Try to load from YAML file first
	configFile := getEnv("CONFIG_FILE", "config.yaml")
	if data, err := ioutil.ReadFile(configFile); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			log.Printf("Error parsing YAML config: %v", err)
		} else {
			log.Printf("Loaded configuration from %s", configFile)
		}
	} else {
		log.Printf("No config file found at %s, using defaults", configFile)
	}

	// Override with environment variables and set defaults
	config.populateDefaults()

	return config
}

// populateDefaults sets default values and applies environment variable overrides
func (c *Config) populateDefaults() {
	// Server defaults
	if c.Server.Host == "" {
		c.Server.Host = getEnv("HOST", "0.0.0.0")
	}
	if c.Server.Port == 0 {
		c.Server.Port = getEnvAsInt("PORT", 8080)
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = getDurationEnv("READ_TIMEOUT", "30s")
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = getDurationEnv("WRITE_TIMEOUT", "30s")
	}

	// CORS defaults
	if len(c.CORS.AllowedMethods) == 0 {
		c.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(c.CORS.AllowedHeaders) == 0 {
		c.CORS.AllowedHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = getEnv("LOG_LEVEL", "info")
	}
	if c.Logging.Format == "" {
		c.Logging.Format = getEnv("LOG_FORMAT", "json")
	}

	// Services defaults
	if c.Services == nil {
		c.Services = make(map[string]ServiceConfig)
	}
	if _, exists := c.Services["analytics"]; !exists {
		c.Services["analytics"] = ServiceConfig{
			URL:     getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8000"),
			Timeout: getDurationEnv("ANALYTICS_SERVICE_TIMEOUT", "10s"),
		}
	}
	if _, exists := c.Services["auth"]; !exists {
		c.Services["auth"] = ServiceConfig{
			URL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
			Timeout: getDurationEnv("AUTH_SERVICE_TIMEOUT", "10s"),
		}
	}
	if _, exists := c.Services["data_management"]; !exists {
		c.Services["data_management"] = ServiceConfig{
			URL:     getEnv("DATA_MANAGEMENT_SERVICE_URL", "http://localhost:8002"),
			Timeout: getDurationEnv("DATA_MANAGEMENT_SERVICE_TIMEOUT", "10s"),
		}
	}
	if _, exists := c.Services["plant_management"]; !exists {
		c.Services["plant_management"] = ServiceConfig{
			URL:     getEnv("PLANT_MANAGEMENT_SERVICE_URL", "http://localhost:8003"),
			Timeout: getDurationEnv("PLANT_MANAGEMENT_SERVICE_TIMEOUT", "10s"),
		}
	}

	// Auth defaults
	if c.Auth.APIKeyHeader == "" {
		c.Auth.APIKeyHeader = getEnv("API_KEY_HEADER", "X-API-Key")
	}
	if c.Auth.JWTSecret == "" {
		c.Auth.JWTSecret = getEnv("JWT_SECRET", "your-secret-key")
	}
	if c.Auth.JWTExpiration == 0 {
		c.Auth.JWTExpiration = getDurationEnv("JWT_EXPIRATION", "24h")
	}

	// Legacy fields for backward compatibility
	c.Port = fmt.Sprintf("%d", c.Server.Port)
	c.GinMode = getEnv("GIN_MODE", "debug")
	c.GraphQLPlaygroundEnabled = getEnvAsBool("GRAPHQL_PLAYGROUND_ENABLED", true)
	c.GraphQLIntrospectionEnabled = getEnvAsBool("GRAPHQL_INTROSPECTION_ENABLED", true)
	c.CORSAllowAllOrigins = c.CORS.AllowAllOrigins
	if !c.CORSAllowAllOrigins {
		c.CORSAllowAllOrigins = getEnvAsBool("CORS_ALLOW_ALL_ORIGINS", true)
	}
	c.LogLevel = c.Logging.Level
	c.LogFormat = c.Logging.Format
	c.AnalyticsServiceURL = c.Services["analytics"].URL
	c.AuthServiceURL = c.Services["auth"].URL
	c.DataManagementServiceURL = c.Services["data_management"].URL
	c.PlantManagementServiceURL = c.Services["plant_management"].URL
}

// GetServiceURL returns the URL for a given service
func (c *Config) GetServiceURL(serviceName string) string {
	if service, exists := c.Services[serviceName]; exists {
		return service.URL
	}
	return ""
}

// GetServiceTimeout returns the timeout for a given service
func (c *Config) GetServiceTimeout(serviceName string) time.Duration {
	if service, exists := c.Services[serviceName]; exists {
		return service.Timeout
	}
	return 10 * time.Second
}

// getEnv gets an environment variable with a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(name string, defaultVal int) int {
	valStr := getEnv(name, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

// getEnvAsBool gets an environment variable as boolean with a default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

// getDurationEnv gets an environment variable as duration with a default value
func getDurationEnv(name string, defaultVal string) time.Duration {
	valStr := getEnv(name, defaultVal)
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	if defaultDuration, err := time.ParseDuration(defaultVal); err == nil {
		return defaultDuration
	}
	return 30 * time.Second
}