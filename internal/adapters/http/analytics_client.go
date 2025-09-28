package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/domain"
)

// AnalyticsHTTPClient implements the AnalyticsClient interface using HTTP calls
type AnalyticsHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAnalyticsHTTPClient creates a new analytics HTTP client
func NewAnalyticsHTTPClient(baseURL string) ports.AnalyticsClient {
	return &AnalyticsHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetSingleMetricReport retrieves a single metric report from analytics service
func (c *AnalyticsHTTPClient) GetSingleMetricReport(ctx context.Context, metricName string, controllerID string, filter domain.AnalyticsFilter) (*domain.AnalyticsReport, error) {
	// Build query parameters
	params := url.Values{}
	params.Add("metric_name", metricName)
	params.Add("id_controlador", controllerID)

	if filter.StartTime != nil {
		params.Add("start_time", filter.StartTime.Format(time.RFC3339))
	}
	if filter.EndTime != nil {
		params.Add("end_time", filter.EndTime.Format(time.RFC3339))
	}
	if filter.Limit != nil {
		params.Add("limit", strconv.Itoa(*filter.Limit))
	}

	url := fmt.Sprintf("%s/analytics/reports/single?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var report domain.AnalyticsReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &report, nil
}

// GetMultiMetricReport retrieves multiple metric reports from analytics service
func (c *AnalyticsHTTPClient) GetMultiMetricReport(ctx context.Context, request domain.MultiMetricReportRequest) (*domain.MultiReportResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/analytics/reports/multi-report", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response domain.MultiReportResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetTrendAnalysis retrieves trend analysis data from analytics service
func (c *AnalyticsHTTPClient) GetTrendAnalysis(ctx context.Context, request domain.TrendAnalysisRequest) (*domain.TrendAnalysis, error) {
	// Build query parameters for trend analysis
	params := url.Values{}
	params.Add("id_controlador", request.ControllerID)
	if request.Filters.StartTime != nil {
		params.Add("start_time", request.Filters.StartTime.Format(time.RFC3339))
	}
	if request.Filters.EndTime != nil {
		params.Add("end_time", request.Filters.EndTime.Format(time.RFC3339))
	}
	params.Add("interval", request.Interval)

	url := fmt.Sprintf("%s/analytics/trends/%s?%s", c.baseURL, request.MetricName, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var analysis domain.TrendAnalysis
	if err := json.Unmarshal(body, &analysis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &analysis, nil
}

// GetSupportedMetrics retrieves the list of supported metrics from analytics service
func (c *AnalyticsHTTPClient) GetSupportedMetrics(ctx context.Context) (*domain.SupportedMetrics, error) {
	url := fmt.Sprintf("%s/analytics/metrics", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// The backend returns a list of strings, but we need to convert to our domain model
	var metricNames []string
	if err := json.Unmarshal(body, &metricNames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to our domain model
	var metrics []domain.MetricInfo
	for _, name := range metricNames {
		metrics = append(metrics, domain.MetricInfo{
			Name:        name,
			Description: fmt.Sprintf("Analytics metric: %s", name),
			Unit:        "various", // Backend doesn't provide unit info
		})
	}

	return &domain.SupportedMetrics{
		Metrics:     metrics,
		TotalCount:  len(metrics),
		RetrievedAt: time.Now(),
	}, nil
}

// GetAnalyticsHealth checks the health of the analytics service
func (c *AnalyticsHTTPClient) GetAnalyticsHealth(ctx context.Context) (*domain.HealthCheck, error) {
	url := fmt.Sprintf("%s/analytics/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Backend returns: {"status": "healthy", "service": "analytics", "timestamp": "..."}
	var healthResponse struct {
		Status    string    `json:"status"`
		Service   string    `json:"service"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(body, &healthResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &domain.HealthCheck{
		Status:    healthResponse.Status,
		CheckedAt: healthResponse.Timestamp,
		Version:   "1.0.0", // Backend doesn't provide version
		Dependencies: map[string]string{
			"analytics_service": healthResponse.Status,
		},
	}, nil
}
