package services

import (
	"context"
	"fmt"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/domain"
)

// AnalyticsService handles business logic for analytics operations
type AnalyticsService struct {
	analyticsClient ports.AnalyticsClient
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(analyticsClient ports.AnalyticsClient) *AnalyticsService {
	return &AnalyticsService{
		analyticsClient: analyticsClient,
	}
}

// GetSingleMetricReport retrieves a single metric report
func (s *AnalyticsService) GetSingleMetricReport(ctx context.Context, metricName string, controllerID string, filters *domain.AnalyticsFilter) (*domain.AnalyticsReport, error) {
	// Create the filter with default values if nil
	var filter domain.AnalyticsFilter
	if filters != nil {
		filter = *filters
	}
	
	report, err := s.analyticsClient.GetSingleMetricReport(ctx, metricName, controllerID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get single metric report: %w", err)
	}
	return report, nil
}

// GetMultiMetricReport retrieves multiple metric reports
func (s *AnalyticsService) GetMultiMetricReport(ctx context.Context, request domain.MultiMetricReportRequest) (*domain.MultiReportResponse, error) {
	response, err := s.analyticsClient.GetMultiMetricReport(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get multi metric report: %w", err)
	}
	return response, nil
}

// GetTrendAnalysis retrieves trend analysis data
func (s *AnalyticsService) GetTrendAnalysis(ctx context.Context, request domain.TrendAnalysisRequest) (*domain.TrendAnalysis, error) {
	analysis, err := s.analyticsClient.GetTrendAnalysis(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get trend analysis: %w", err)
	}
	return analysis, nil
}

// GetSupportedMetrics retrieves the list of supported metrics
func (s *AnalyticsService) GetSupportedMetrics(ctx context.Context) (*domain.SupportedMetrics, error) {
	metrics, err := s.analyticsClient.GetSupportedMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get supported metrics: %w", err)
	}
	return metrics, nil
}

// GetAnalyticsHealth checks the health of the analytics service
func (s *AnalyticsService) GetAnalyticsHealth(ctx context.Context) (*domain.HealthCheck, error) {
	health, err := s.analyticsClient.GetAnalyticsHealth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics health: %w", err)
	}
	return health, nil
}