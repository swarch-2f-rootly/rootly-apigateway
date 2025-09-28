package ports

import (
	"context"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/domain"
)

// AnalyticsService defines the business logic interface for analytics operations
type AnalyticsService interface {
	// GetSingleMetricReport retrieves a single metric report
	GetSingleMetricReport(ctx context.Context, metricName string, controllerID string, filters *domain.AnalyticsFilter) (*domain.AnalyticsReport, error)

	// GetMultiMetricReport retrieves multiple metric reports
	GetMultiMetricReport(ctx context.Context, request domain.MultiMetricReportRequest) (*domain.MultiReportResponse, error)

	// GetTrendAnalysis retrieves trend analysis data
	GetTrendAnalysis(ctx context.Context, request domain.TrendAnalysisRequest) (*domain.TrendAnalysis, error)

	// GetSupportedMetrics retrieves the list of supported metrics
	GetSupportedMetrics(ctx context.Context) (*domain.SupportedMetrics, error)

	// GetAnalyticsHealth checks the health of the analytics service
	GetAnalyticsHealth(ctx context.Context) (*domain.HealthCheck, error)
}