package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/adapters/graphql"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/adapters/graphql/generated"
	httpAdapter "github.com/swarch-2f-rootly/rootly-apigateway/internal/adapters/http"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/config"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize dependencies using dependency injection
	analyticsClient := httpAdapter.NewAnalyticsHTTPClient(cfg.AnalyticsServiceURL)
	analyticsService := services.NewAnalyticsService(analyticsClient)
	
	// Initialize Gin router
	gin.SetMode(cfg.GinMode)
	router := gin.New()

	// Add middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup CORS
	if cfg.CORSAllowAllOrigins {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
		corsConfig.AllowMethods = []string{"GET", "POST", "OPTIONS"}
		router.Use(cors.New(corsConfig))
	}

	// Setup GraphQL with all dependencies
	setupGraphQL(router, cfg, analyticsService)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "rootly-apigateway",
			"version": "1.0.0",
		})
	})

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("API Gateway starting on port %s", cfg.Port)
	
	if cfg.GraphQLPlaygroundEnabled {
		log.Printf("GraphQL Playground available at http://localhost:%s/playground", cfg.Port)
	}
	
	log.Printf("GraphQL endpoint available at http://localhost:%s/graphql", cfg.Port)
	log.Printf("Health check available at http://localhost:%s/health", cfg.Port)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupGraphQL configures GraphQL handler and playground with all dependencies
func setupGraphQL(router *gin.Engine, cfg *config.Config, analyticsService *services.AnalyticsService) {
	// Create GraphQL resolver with injected dependencies
	resolver := graphql.NewResolver(analyticsService)

	// Create GraphQL server configuration
	graphqlConfig := generated.Config{
		Resolvers: resolver,
	}

	// Create GraphQL server with configuration
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(graphqlConfig))

	// Configure GraphQL server options
	srv.AddTransport(transport.Websocket{KeepAlivePingInterval: 10 * time.Second})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Add recovery middleware
	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		log.Printf("GraphQL panic recovered: %v", err)
		return fmt.Errorf("internal server error")
	})

	// Enable introspection if configured
	if cfg.GraphQLIntrospectionEnabled {
		srv.Use(extension.Introspection{})
	}

	// GraphQL endpoint - accepts both GET and POST
	router.POST("/graphql", ginHandler(srv))
	router.GET("/graphql", ginHandler(srv))

	// GraphQL Playground (only if enabled)
	if cfg.GraphQLPlaygroundEnabled {
		playgroundHandler := playground.Handler("GraphQL Playground", "/graphql")
		router.GET("/playground", ginHandler(playgroundHandler))
		log.Printf("GraphQL Playground enabled at /playground")
	}

	log.Printf("GraphQL server configured with analytics service")
}

// ginHandler converts http.Handler to gin.HandlerFunc
func ginHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}