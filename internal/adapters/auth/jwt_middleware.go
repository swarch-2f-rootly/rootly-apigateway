package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
)

// JWTMiddleware handles JWT token validation against the auth service
type JWTMiddleware struct {
	authServiceURL     string
	validationEndpoint string
	validationStrategy string
	httpClient         *http.Client
	logger             ports.Logger
	configProvider     ports.ConfigProvider
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(
	authServiceURL string,
	validationEndpoint string,
	validationStrategy string,
	logger ports.Logger,
	configProvider ports.ConfigProvider,
) *JWTMiddleware {
	return &JWTMiddleware{
		authServiceURL:     authServiceURL,
		validationEndpoint: validationEndpoint,
		validationStrategy: validationStrategy,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger:         logger,
		configProvider: configProvider,
	}
}

// ValidateRequest validates JWT token for protected routes
func (m *JWTMiddleware) ValidateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Always allow OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get route configuration
		routeConfig, found := m.configProvider.GetRouteConfig(c.Request.URL.Path, c.Request.Method)
		
		// If route not found or auth not required, skip validation
		if !found || !routeConfig.AuthRequired {
			m.logger.Debug("Route does not require authentication", map[string]interface{}{
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
				"found":  found,
			})
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing Authorization header", map[string]interface{}{
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			})
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			m.logger.Warn("Invalid Authorization header format", map[string]interface{}{
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			})
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token against auth service
		user, err := m.validateToken(c.Request.Context(), token)
		if err != nil {
			m.logger.Warn("Token validation failed", map[string]interface{}{
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
				"error":  err.Error(),
			})
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		
		m.logger.Debug("Token validated successfully", map[string]interface{}{
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
			"user_id": user.ID,
		})

		c.Next()
	}
}

// TokenValidationRequest represents the request to validate a token
type TokenValidationRequest struct {
	Token string `json:"token"`
}

// TokenValidationResponse represents the response from token validation
type TokenValidationResponse struct {
	Valid   bool              `json:"valid"`
	UserID  string            `json:"user_id,omitempty"`
	Email   string            `json:"email,omitempty"`
	Roles   []string          `json:"roles,omitempty"`
	Message string            `json:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UserInfo represents user information from the auth service
type UserInfo struct {
	ID       string                 `json:"id"`
	Email    string                 `json:"email"`
	Username string                 `json:"username,omitempty"`
	Roles    []string               `json:"roles,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// validateToken validates a JWT token against the auth service
func (m *JWTMiddleware) validateToken(ctx context.Context, token string) (*UserInfo, error) {
	// Prepare validation request
	validationReq := TokenValidationRequest{
		Token: token,
	}

	reqBody, err := json.Marshal(validationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validation request: %w", err)
	}

	// Create HTTP request to auth service
	validateURL := fmt.Sprintf("%s%s", m.authServiceURL, m.validationEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", validateURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	m.logger.Debug("Validating token against auth service", map[string]interface{}{
		"validation_url":      validateURL,
		"validation_strategy": m.validationStrategy,
	})

	// Send request to auth service
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token with auth service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read validation response: %w", err)
	}

	// Check if token is valid
	if resp.StatusCode != http.StatusOK {
		m.logger.Debug("Token validation failed", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
		})
		return nil, fmt.Errorf("token validation failed with status %d", resp.StatusCode)
	}

	// Parse response
	var validationResp TokenValidationResponse
	if err := json.Unmarshal(body, &validationResp); err != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	if !validationResp.Valid {
		return nil, fmt.Errorf("token is not valid: %s", validationResp.Message)
	}

	// Return user information
	return &UserInfo{
		ID:       validationResp.UserID,
		Email:    validationResp.Email,
		Roles:    validationResp.Roles,
		Metadata: validationResp.Metadata,
	}, nil
}
