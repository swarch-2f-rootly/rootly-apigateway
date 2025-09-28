package graphql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/adapters/graphql/generated"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/domain"
)

type Resolver struct {
	analyticsService ports.AnalyticsService
}

// NewResolver creates a new resolver with dependencies
func NewResolver(analyticsService ports.AnalyticsService) *Resolver {
	return &Resolver{
		analyticsService: analyticsService,
	}
}

// Service is the resolver for the service field.
func (r *analyticsHealthStatusResolver) Service(ctx context.Context, obj *domain.AnalyticsHealthStatus) (string, error) {
	if status, exists := obj.Dependencies["analytics_service"]; exists {
		return status, nil
	}
	return "unknown", nil
}

// Influxdb is the resolver for the influxdb field.
func (r *analyticsHealthStatusResolver) Influxdb(ctx context.Context, obj *domain.AnalyticsHealthStatus) (string, error) {
	if status, exists := obj.Dependencies["influxdb"]; exists {
		return status, nil
	}
	return "unknown", nil
}

// InfluxdbURL is the resolver for the influxdbUrl field.
func (r *analyticsHealthStatusResolver) InfluxdbURL(ctx context.Context, obj *domain.AnalyticsHealthStatus) (string, error) {
	if url, exists := obj.Dependencies["influxdb_url"]; exists {
		return url, nil
	}
	return "", nil
}

// Timestamp is the resolver for the timestamp field.
func (r *analyticsHealthStatusResolver) Timestamp(ctx context.Context, obj *domain.AnalyticsHealthStatus) (*time.Time, error) {
	return &obj.CheckedAt, nil
}

// MicrocontrollerID is the resolver for the microcontrollerId field.
func (r *analyticsReportResolver) MicrocontrollerID(ctx context.Context, obj *domain.AnalyticsReport) (*uuid.UUID, error) {
	// Parse the controller ID string to UUID
	id, err := uuid.Parse(obj.ControllerID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Filters is the resolver for the filters field.
func (r *analyticsReportResolver) Filters(ctx context.Context, obj *domain.AnalyticsReport) (*domain.AnalyticsFilter, error) {
	return &obj.FiltersApplied, nil
}

// MicrocontrollerID is the resolver for the microcontrollerId field.
func (r *metricResultResolver) MicrocontrollerID(ctx context.Context, obj *domain.MetricResult) (*uuid.UUID, error) {
	// Parse the controller ID string to UUID
	id, err := uuid.Parse(obj.ControllerID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Reports is the resolver for the reports field.
func (r *multiMetricReportResolver) Reports(ctx context.Context, obj *domain.MultiReportResponse) ([]*domain.AnalyticsReport, error) {
	// Convert map of reports to slice
	reports := make([]*domain.AnalyticsReport, 0, len(obj.Reports))
	for _, report := range obj.Reports {
		reportCopy := report // Create a copy to avoid pointer issues
		reports = append(reports, &reportCopy)
	}
	return reports, nil
}

// GetSingleMetricReport is the resolver for the getSingleMetricReport field.
func (r *queryResolver) GetSingleMetricReport(ctx context.Context, metricName string, controllerID string, filters *generated.AnalyticsFilterInput) (*domain.AnalyticsReport, error) {
	// Convert GraphQL input to domain filters
	var domainFilters *domain.AnalyticsFilter
	if filters != nil {
		domainFilters = &domain.AnalyticsFilter{
			StartTime: filters.StartTime,
			EndTime:   filters.EndTime,
			Limit:     filters.Limit,
		}
	}
	
	return r.analyticsService.GetSingleMetricReport(ctx, metricName, controllerID, domainFilters)
}

// GetMultiMetricReport is the resolver for the getMultiMetricReport field.
func (r *queryResolver) GetMultiMetricReport(ctx context.Context, input domain.MultiMetricReportRequest) (*domain.MultiReportResponse, error) {
	return r.analyticsService.GetMultiMetricReport(ctx, input)
}

// GetTrendAnalysis is the resolver for the getTrendAnalysis field.
func (r *queryResolver) GetTrendAnalysis(ctx context.Context, input domain.TrendAnalysisRequest) (*domain.TrendAnalysis, error) {
	return r.analyticsService.GetTrendAnalysis(ctx, input)
}

// GetSupportedMetrics is the resolver for the getSupportedMetrics field.
func (r *queryResolver) GetSupportedMetrics(ctx context.Context) ([]string, error) {
	supportedMetrics, err := r.analyticsService.GetSupportedMetrics(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convert to simple string slice
	metrics := make([]string, len(supportedMetrics.Metrics))
	for i, metric := range supportedMetrics.Metrics {
		metrics[i] = metric.Name
	}
	
	return metrics, nil
}

// GetAnalyticsHealth is the resolver for the getAnalyticsHealth field.
func (r *queryResolver) GetAnalyticsHealth(ctx context.Context) (*domain.AnalyticsHealthStatus, error) {
	healthCheck, err := r.analyticsService.GetAnalyticsHealth(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convert HealthCheck to AnalyticsHealthStatus
	return &domain.AnalyticsHealthStatus{
		ServiceName:  "analytics",
		Status:       healthCheck.Status,
		CheckedAt:    healthCheck.CheckedAt,
		Version:      healthCheck.Version,
		Dependencies: healthCheck.Dependencies,
	}, nil
}

// MicrocontrollerID is the resolver for the microcontrollerId field.
func (r *trendAnalysisResolver) MicrocontrollerID(ctx context.Context, obj *domain.TrendAnalysis) (*uuid.UUID, error) {
	// Parse the controller ID string to UUID
	id, err := uuid.Parse(obj.ControllerID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Filters is the resolver for the filters field.
func (r *trendAnalysisResolver) Filters(ctx context.Context, obj *domain.TrendAnalysis) (*domain.AnalyticsFilter, error) {
	return &obj.FiltersApplied, nil
}

// TotalPoints is the resolver for the totalPoints field.
func (r *trendAnalysisResolver) TotalPoints(ctx context.Context, obj *domain.TrendAnalysis) (int, error) {
	return len(obj.DataPoints), nil
}

// AverageValue is the resolver for the averageValue field.
func (r *trendAnalysisResolver) AverageValue(ctx context.Context, obj *domain.TrendAnalysis) (float64, error) {
	if len(obj.DataPoints) == 0 {
		return 0.0, nil
	}
	
	sum := 0.0
	for _, point := range obj.DataPoints {
		sum += point.Value
	}
	return sum / float64(len(obj.DataPoints)), nil
}

// MinValue is the resolver for the minValue field.
func (r *trendAnalysisResolver) MinValue(ctx context.Context, obj *domain.TrendAnalysis) (float64, error) {
	if len(obj.DataPoints) == 0 {
		return 0.0, nil
	}
	
	min := obj.DataPoints[0].Value
	for _, point := range obj.DataPoints {
		if point.Value < min {
			min = point.Value
		}
	}
	return min, nil
}

// MaxValue is the resolver for the maxValue field.
func (r *trendAnalysisResolver) MaxValue(ctx context.Context, obj *domain.TrendAnalysis) (float64, error) {
	if len(obj.DataPoints) == 0 {
		return 0.0, nil
	}
	
	max := obj.DataPoints[0].Value
	for _, point := range obj.DataPoints {
		if point.Value > max {
			max = point.Value
		}
	}
	return max, nil
}

// AnalyticsHealthStatus returns generated.AnalyticsHealthStatusResolver implementation.
func (r *Resolver) AnalyticsHealthStatus() generated.AnalyticsHealthStatusResolver {
	return &analyticsHealthStatusResolver{r}
}

// AnalyticsReport returns generated.AnalyticsReportResolver implementation.
func (r *Resolver) AnalyticsReport() generated.AnalyticsReportResolver {
	return &analyticsReportResolver{r}
}

// MetricResult returns generated.MetricResultResolver implementation.
func (r *Resolver) MetricResult() generated.MetricResultResolver { return &metricResultResolver{r} }

// MultiMetricReport returns generated.MultiMetricReportResolver implementation.
func (r *Resolver) MultiMetricReport() generated.MultiMetricReportResolver {
	return &multiMetricReportResolver{r}
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// TrendAnalysis returns generated.TrendAnalysisResolver implementation.
func (r *Resolver) TrendAnalysis() generated.TrendAnalysisResolver { return &trendAnalysisResolver{r} }

// MultiMetricReportInput returns generated.MultiMetricReportInputResolver implementation.
func (r *Resolver) MultiMetricReportInput() generated.MultiMetricReportInputResolver {
	return &multiMetricReportInputResolver{r}
}

// TrendAnalysisInput returns generated.TrendAnalysisInputResolver implementation.
func (r *Resolver) TrendAnalysisInput() generated.TrendAnalysisInputResolver {
	return &trendAnalysisInputResolver{r}
}

// Filters is the resolver for input filters field.
func (r *multiMetricReportInputResolver) Filters(ctx context.Context, obj *domain.MultiMetricReportRequest, data *generated.AnalyticsFilterInput) error {
	// This is called during input processing - typically no implementation needed
	return nil
}

// StartTime is the resolver for input startTime field.
func (r *trendAnalysisInputResolver) StartTime(ctx context.Context, obj *domain.TrendAnalysisRequest, data *time.Time) error {
	// This is called during input processing - typically no implementation needed
	return nil
}

// EndTime is the resolver for input endTime field.
func (r *trendAnalysisInputResolver) EndTime(ctx context.Context, obj *domain.TrendAnalysisRequest, data *time.Time) error {
	// This is called during input processing - typically no implementation needed
	return nil
}

type analyticsHealthStatusResolver struct{ *Resolver }
type analyticsReportResolver struct{ *Resolver }
type metricResultResolver struct{ *Resolver }
type multiMetricReportResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type trendAnalysisResolver struct{ *Resolver }
type multiMetricReportInputResolver struct{ *Resolver }
type trendAnalysisInputResolver struct{ *Resolver }
