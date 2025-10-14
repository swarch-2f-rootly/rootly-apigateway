package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
)

// AuthService implements authentication functionality
type AuthService struct {
	jwtSecret     string
	jwtExpiration time.Duration
	apiKeys       map[string]bool // In production, this would be a database
	logger        ports.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(jwtSecret string, jwtExpiration time.Duration, logger ports.Logger) *AuthService {
	// Initialize with some default API keys for testing
	apiKeys := map[string]bool{
		"rootly-api-key-123":     true,
		"test-api-key":           true,
		"dashboard-api-key":      true,
		"analytics-service-key":  true,
	}

	return &AuthService{
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
		apiKeys:       apiKeys,
		logger:        logger,
	}
}

// ValidateAPIKey validates an API key
func (as *AuthService) ValidateAPIKey(ctx context.Context, apiKey string) (bool, error) {
	if apiKey == "" {
		return false, errors.New("API key is empty")
	}

	valid, exists := as.apiKeys[apiKey]
	if !exists {
		as.logger.Warn("Invalid API key used", map[string]interface{}{
			"api_key_prefix": as.maskAPIKey(apiKey),
		})
		return false, nil
	}

	as.logger.Debug("API key validated", map[string]interface{}{
		"api_key_prefix": as.maskAPIKey(apiKey),
		"valid":          valid,
	})

	return valid, nil
}

// ValidateJWT validates a JWT token and returns user information
func (as *AuthService) ValidateJWT(ctx context.Context, tokenString string) (*ports.UserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(as.jwtSecret), nil
	})

	if err != nil {
		as.logger.Warn("JWT validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("JWT token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid JWT claims")
	}

	// Extract user information from claims
	userInfo := &ports.UserInfo{
		Metadata: make(map[string]interface{}),
	}

	if id, ok := claims["sub"].(string); ok {
		userInfo.ID = id
	}

	if username, ok := claims["username"].(string); ok {
		userInfo.Username = username
	}

	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}

	if rolesInterface, ok := claims["roles"].([]interface{}); ok {
		roles := make([]string, len(rolesInterface))
		for i, role := range rolesInterface {
			if roleStr, ok := role.(string); ok {
				roles[i] = roleStr
			}
		}
		userInfo.Roles = roles
	}

	// Add any additional metadata
	for key, value := range claims {
		if key != "sub" && key != "username" && key != "email" && key != "roles" && key != "exp" && key != "iat" {
			userInfo.Metadata[key] = value
		}
	}

	as.logger.Debug("JWT validated successfully", map[string]interface{}{
		"user_id":  userInfo.ID,
		"username": userInfo.Username,
		"roles":    userInfo.Roles,
	})

	return userInfo, nil
}

// GenerateJWT generates a JWT token for the given user information
func (as *AuthService) GenerateJWT(ctx context.Context, userInfo *ports.UserInfo) (string, error) {
	// Create claims
	claims := jwt.MapClaims{
		"sub":      userInfo.ID,
		"username": userInfo.Username,
		"email":    userInfo.Email,
		"roles":    userInfo.Roles,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(as.jwtExpiration).Unix(),
	}

	// Add metadata to claims
	for key, value := range userInfo.Metadata {
		claims[key] = value
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(as.jwtSecret))
	if err != nil {
		as.logger.Error("Failed to generate JWT", err, map[string]interface{}{
			"user_id": userInfo.ID,
		})
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	as.logger.Info("JWT generated successfully", map[string]interface{}{
		"user_id":  userInfo.ID,
		"username": userInfo.Username,
		"expires":  time.Now().Add(as.jwtExpiration).Format(time.RFC3339),
	})

	return tokenString, nil
}

// AddAPIKey adds a new API key (for testing purposes)
func (as *AuthService) AddAPIKey(apiKey string) {
	as.apiKeys[apiKey] = true
	as.logger.Info("API key added", map[string]interface{}{
		"api_key_prefix": as.maskAPIKey(apiKey),
	})
}

// RemoveAPIKey removes an API key
func (as *AuthService) RemoveAPIKey(apiKey string) {
	delete(as.apiKeys, apiKey)
	as.logger.Info("API key removed", map[string]interface{}{
		"api_key_prefix": as.maskAPIKey(apiKey),
	})
}

// maskAPIKey masks an API key for logging purposes
func (as *AuthService) maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
}

// RefreshJWT refreshes a JWT token if it's still valid but close to expiration
func (as *AuthService) RefreshJWT(ctx context.Context, tokenString string) (string, error) {
	// Parse the token without validation to check expiration
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(as.jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token for refresh: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	// Check if token is close to expiration (within 1 hour)
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if time.Until(expTime) > time.Hour {
			return tokenString, nil // Token is still valid for more than 1 hour
		}
	}

	// Extract user info and generate new token
	userInfo := &ports.UserInfo{
		ID:       getStringClaim(claims, "sub"),
		Username: getStringClaim(claims, "username"),
		Email:    getStringClaim(claims, "email"),
		Metadata: make(map[string]interface{}),
	}

	if rolesInterface, ok := claims["roles"].([]interface{}); ok {
		roles := make([]string, len(rolesInterface))
		for i, role := range rolesInterface {
			if roleStr, ok := role.(string); ok {
				roles[i] = roleStr
			}
		}
		userInfo.Roles = roles
	}

	return as.GenerateJWT(ctx, userInfo)
}

// getStringClaim safely extracts a string claim from JWT claims
func getStringClaim(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}