package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/adapters/auth"
	httpAdapter "github.com/swarch-2f-rootly/rootly-apigateway/internal/adapters/http"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/adapters/logger"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/config"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/services"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/services/strategies"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	logger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format, "api-gateway")

	// Log JWT configuration for debugging
	jwtSecretPreview := "NOT_SET"
	if cfg.Auth.JWTSecret != "" {
		if len(cfg.Auth.JWTSecret) > 20 {
			jwtSecretPreview = cfg.Auth.JWTSecret[:20] + "..."
		} else {
			jwtSecretPreview = cfg.Auth.JWTSecret
		}
	}

	logger.Info("Starting Rootly API Gateway", map[string]interface{}{
		"version": "1.0.0",
		"port":    cfg.Server.Port,
		"routes":  len(cfg.Routes),
	})

	logger.Info("🔐 JWT Configuration", map[string]interface{}{
		"jwt_secret_preview": jwtSecretPreview,
		"jwt_secret_length":  len(cfg.Auth.JWTSecret),
		"jwt_expiration":     cfg.Auth.JWTExpiration,
		"api_key_header":     cfg.Auth.APIKeyHeader,
	})

	// Initialize HTTP client
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Initialize auth service
	authService := auth.NewAuthService(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTExpiration,
		logger,
	)

	// Initialize config provider
	configProvider := httpAdapter.NewConfigProvider(cfg, logger)

	// Initialize strategy manager
	strategyManager := services.NewStrategyManager(logger)

	// Register strategies
	registerStrategies(strategyManager, logger)

	// Initialize gateway service
	gatewayService := services.NewGatewayService(
		strategyManager,
		nil, // Service orchestrator - could be implemented separately
		authService,
		logger,
		httpClient,
		configProvider,
	)

	// Initialize HTTP handler
	gatewayHandler := httpAdapter.NewGatewayHandler(
		gatewayService,
		configProvider,
		logger,
	)

	// Setup Gin router
	gin.SetMode(func() string {
		if cfg.Logging.Level == "debug" {
			return gin.DebugMode
		}
		return gin.ReleaseMode
	}())

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup CORS - MUST be before JWT middleware to handle preflight requests
	corsConfig := cors.DefaultConfig()
	if cfg.CORS.AllowAllOrigins {
		corsConfig.AllowAllOrigins = true
	} else if len(cfg.CORS.AllowedOrigins) > 0 {
		corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	} else {
		// Default fallback if no CORS config is set
		corsConfig.AllowAllOrigins = true
	}
	corsConfig.AllowHeaders = append(cfg.CORS.AllowedHeaders, "Accept", "Accept-Language", "Content-Language", "X-Request-ID")
	corsConfig.AllowMethods = cfg.CORS.AllowedMethods
	corsConfig.AllowCredentials = true
	corsConfig.ExposeHeaders = []string{"Content-Length", "Content-Type", "Authorization"}
	corsConfig.MaxAge = 12 * time.Hour
	
	router.Use(cors.New(corsConfig))

	logger.Info("CORS middleware configured", map[string]interface{}{
		"allow_all_origins":  corsConfig.AllowAllOrigins,
		"allowed_origins":    corsConfig.AllowOrigins,
		"allowed_methods":    corsConfig.AllowMethods,
		"allowed_headers":    corsConfig.AllowHeaders,
		"allow_credentials":  corsConfig.AllowCredentials,
	})

	// Setup JWT middleware for authentication
	jwtMiddleware := auth.NewJWTMiddleware(
		cfg.Services["auth"].URL,
		cfg.Auth.ValidationEndpoint,
		cfg.Auth.ValidationStrategy,
		logger,
		configProvider,
	)
	router.Use(jwtMiddleware.ValidateRequest())

	logger.Info("JWT middleware configured", map[string]interface{}{
		"auth_service_url":     cfg.Services["auth"].URL,
		"validation_endpoint":  cfg.Auth.ValidationEndpoint,
		"validation_strategy":  cfg.Auth.ValidationStrategy,
		"jwt_expiration":       cfg.Auth.JWTExpiration,
	})

	// Register routes
	gatewayHandler.RegisterRoutes(router)

	// Setup server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", map[string]interface{}{
			"address": server.Addr,
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server startup failed", err, map[string]interface{}{
				"address": server.Addr,
			})
			os.Exit(1)
		}
	}()

	logger.Info("API Gateway started successfully", map[string]interface{}{
		"address":    server.Addr,
		"routes":     len(cfg.Routes),
		"services":   len(cfg.Services),
		"strategies": len(strategyManager.ListStrategies()),
	})

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...", nil)

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", err, nil)
		os.Exit(1)
	}

	logger.Info("Server exited", nil)
}

// registerStrategies registers all available strategies
func registerStrategies(strategyManager *services.StrategyManager, logger ports.Logger) {
	// Register proxy strategy
	proxyStrategy := strategies.NewProxyStrategy()
	strategyManager.RegisterStrategy(proxyStrategy.GetName(), proxyStrategy)

	// Register business logic strategies
	dashboardStrategy := strategies.NewDashboardOrchestratorStrategy()
	strategyManager.RegisterStrategy(dashboardStrategy.GetName(), dashboardStrategy)

	plantReportStrategy := strategies.NewPlantFullReportStrategy()
	strategyManager.RegisterStrategy(plantReportStrategy.GetName(), plantReportStrategy)

	// Register GraphQL strategies
	localSchemaStrategy := strategies.NewLocalSchemaStrategy()
	strategyManager.RegisterStrategy(localSchemaStrategy.GetName(), localSchemaStrategy)

	proxyGraphQLStrategy := strategies.NewGraphQLProxyStrategy()
	strategyManager.RegisterStrategy("graphql_proxy", proxyGraphQLStrategy)

	logger.Info("Strategies registered", map[string]interface{}{
		"strategies": strategyManager.ListStrategies(),
	})
}
