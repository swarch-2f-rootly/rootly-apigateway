package ports

import (
	"context"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/domain"
)

// AnalyticsClient defines the interface for communicating with the analytics service
type AnalyticsClient interface {
	// GetSingleMetricReport retrieves a single metric report from analytics service
	// Maps to: GET /analytics/reports/single?metric_name={name}&id_controlador={id}&...
	GetSingleMetricReport(ctx context.Context, metricName string, controllerID string, filter domain.AnalyticsFilter) (*domain.AnalyticsReport, error)

	// GetMultiMetricReport retrieves multiple metric reports from analytics service
	// Maps to: POST /analytics/reports/multi-report
	GetMultiMetricReport(ctx context.Context, request domain.MultiMetricReportRequest) (*domain.MultiReportResponse, error)

	// GetTrendAnalysis retrieves trend analysis data from analytics service
	// Maps to: GET /analytics/trends/{metric_name}?id_controlador={id}&start_time={start}&end_time={end}&interval={interval}
	GetTrendAnalysis(ctx context.Context, request domain.TrendAnalysisRequest) (*domain.TrendAnalysis, error)

	// GetSupportedMetrics retrieves the list of supported metrics from analytics service
	// Maps to: GET /analytics/metrics
	GetSupportedMetrics(ctx context.Context) (*domain.SupportedMetrics, error)

	// GetAnalyticsHealth checks the health of the analytics service
	// Maps to: GET /analytics/health
	GetAnalyticsHealth(ctx context.Context) (*domain.HealthCheck, error)
}
