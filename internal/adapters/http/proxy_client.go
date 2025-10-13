package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProxyClient handles HTTP proxy requests to backend services
type ProxyClient struct {
	httpClient *http.Client
}

// NewProxyClient creates a new proxy client
func NewProxyClient() *ProxyClient {
	return &ProxyClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest forwards a request to a backend service
func (p *ProxyClient) ProxyRequest(
	ctx context.Context,
	method string,
	targetURL string,
	body io.Reader,
	headers map[string]string,
) (*http.Response, error) {

	// Create request to target service
	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers from original request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Make request to backend service
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}
