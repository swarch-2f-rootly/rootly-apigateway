package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProxyHandlers contains all proxy handlers
type ProxyHandlers struct {
	proxyClient *ProxyClient
	authURL     string
	plantsURL   string
	dataURL     string
}

// NewProxyHandlers creates new proxy handlers
func NewProxyHandlers(authURL, plantsURL, dataURL string) *ProxyHandlers {
	return &ProxyHandlers{
		proxyClient: NewProxyClient(),
		authURL:     authURL,
		plantsURL:   plantsURL,
		dataURL:     dataURL,
	}
}

// ProxyToAuthService proxies requests to authentication service
func (h *ProxyHandlers) ProxyToAuthService(c *gin.Context) {
	// Build target URL:
	// - Special case: map /api/v1/auth/health -> /health (service root)
	// - Otherwise: forward full path (service exposes /api/v1/auth/*)
	var targetURL string
	if c.Param("path") == "/health" {
		targetURL = h.authURL + "/health"
	} else {
		targetURL = h.authURL + c.Request.URL.Path
	}

	// Copy request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// Copy headers
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Make proxy request
	resp, err := h.proxyClient.ProxyRequest(
		c.Request.Context(),
		c.Request.Method,
		targetURL,
		bytes.NewReader(body),
		headers,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to proxy request"})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Copy response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	// Return response with same status code
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// ProxyToPlantService proxies requests to plant management service
func (h *ProxyHandlers) ProxyToPlantService(c *gin.Context) {
	// Forward full path (service exposes /api/v1/plants/*)
	targetURL := h.plantsURL + c.Request.URL.Path

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	resp, err := h.proxyClient.ProxyRequest(
		c.Request.Context(),
		c.Request.Method,
		targetURL,
		bytes.NewReader(body),
		headers,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to proxy to plant service"})
		return
	}
	defer resp.Body.Close()

	// Copy response
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// ProxyToDataService proxies requests to data management service
func (h *ProxyHandlers) ProxyToDataService(c *gin.Context) {
	// Forward full path unless you need a special-case mapping
	targetURL := h.dataURL + c.Request.URL.Path

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	resp, err := h.proxyClient.ProxyRequest(
		c.Request.Context(),
		c.Request.Method,
		targetURL,
		bytes.NewReader(body),
		headers,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to proxy to data service"})
		return
	}
	defer resp.Body.Close()

	// Copy response
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}
