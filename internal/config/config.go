package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the API Gateway
type Config struct {
	// Service URLs
	AnalyticsServiceURL       string
	AuthServiceURL            string
	DataManagementServiceURL  string
	PlantManagementServiceURL string

	// Server configuration
	Port    string
	GinMode string

	// GraphQL configuration
	GraphQLPlaygroundEnabled    bool
	GraphQLIntrospectionEnabled bool

	// CORS configuration
	CORSAllowAllOrigins bool

	// Logging configuration
	LogLevel  string
	LogFormat string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config := &Config{
		// Service URLs
		AnalyticsServiceURL:       getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8000"),
		AuthServiceURL:            getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		DataManagementServiceURL:  getEnv("DATA_MANAGEMENT_SERVICE_URL", "http://localhost:8002"),
		PlantManagementServiceURL: getEnv("PLANT_MANAGEMENT_SERVICE_URL", "http://localhost:8003"),

		// Server configuration
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		// GraphQL configuration
		GraphQLPlaygroundEnabled:    getEnvAsBool("GRAPHQL_PLAYGROUND_ENABLED", true),
		GraphQLIntrospectionEnabled: getEnvAsBool("GRAPHQL_INTROSPECTION_ENABLED", true),

		// CORS configuration
		CORSAllowAllOrigins: getEnvAsBool("CORS_ALLOW_ALL_ORIGINS", true),

		// Logging configuration
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}

	return config
}

// getEnv gets an environment variable with a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
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
