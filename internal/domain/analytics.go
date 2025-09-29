package domain

import (
	"time"
)

// AnalyticsFilter represents filters for analytics queries
type AnalyticsFilter struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     *int       `json:"limit,omitempty"`
}

// MetricResult represents the result of a single metric calculation
type MetricResult struct {
	MetricName   string    `json:"metric_name"`
	Value        float64   `json:"value"`
	Unit         string    `json:"unit"`
	CalculatedAt time.Time `json:"calculated_at"`
	ControllerID string    `json:"controller_id"`
	Description  *string   `json:"description,omitempty"`
}

// AnalyticsReport represents a complete analytics report
type AnalyticsReport struct {
	ControllerID    string          `json:"controller_id"`
	Metrics         []MetricResult  `json:"metrics"`
	GeneratedAt     time.Time       `json:"generated_at"`
	DataPointsCount int             `json:"data_points_count"`
	FiltersApplied  AnalyticsFilter `json:"filters_applied"`
}

// TrendDataPoint represents a single data point in trend analysis
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Interval  string    `json:"interval"`
}

// TrendAnalysis represents trend analysis for a specific metric
type TrendAnalysis struct {
	MetricName     string           `json:"metric_name"`
	ControllerID   string           `json:"controller_id"`
	DataPoints     []TrendDataPoint `json:"data_points"`
	Interval       string           `json:"interval"`
	GeneratedAt    time.Time        `json:"generated_at"`
	FiltersApplied AnalyticsFilter  `json:"filters_applied"`
}

// MultiMetricReportRequest represents request for multi-controller analytics
type MultiMetricReportRequest struct {
	Controllers []string        `json:"controllers"`
	Metrics     []string        `json:"metrics"`
	Filters     AnalyticsFilter `json:"filters"`
}

// MultiReportResponse represents a multi-controller analytics report
type MultiReportResponse struct {
	Reports          map[string]AnalyticsReport `json:"reports"`
	GeneratedAt      time.Time                  `json:"generated_at"`
	TotalControllers int                        `json:"total_controllers"`
	TotalMetrics     int                        `json:"total_metrics"`
}

// TrendAnalysisRequest represents request for trend analysis
type TrendAnalysisRequest struct {
	ControllerID string          `json:"controller_id"`
	MetricName   string          `json:"metric_name"`
	Interval     string          `json:"interval"`
	Filters      AnalyticsFilter `json:"filters"`
}

// SupportedMetrics represents available metrics information
type SupportedMetrics struct {
	Metrics     []MetricInfo `json:"metrics"`
	TotalCount  int          `json:"total_count"`
	RetrievedAt time.Time    `json:"retrieved_at"`
}

// MetricInfo represents information about a supported metric
type MetricInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
}

// HealthCheck represents service health status
type HealthCheck struct {
	Status       string            `json:"status"`
	CheckedAt    time.Time         `json:"checked_at"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

// AnalyticsHealthStatus represents the health status of analytics service
type AnalyticsHealthStatus struct {
	ServiceName  string            `json:"service_name"`
	Status       string            `json:"status"`
	CheckedAt    time.Time         `json:"checked_at"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}
