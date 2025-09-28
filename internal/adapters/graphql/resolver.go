package graphql

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/adapters/graphql/generated"
	"github.com/swarch-2f-rootly/rootly-apigateway/internal/domain"
)

type Resolver struct{}

// Service is the resolver for the service field.
func (r *analyticsHealthStatusResolver) Service(ctx context.Context, obj *domain.AnalyticsHealthStatus) (string, error) {
	panic("not implemented")
}

// Influxdb is the resolver for the influxdb field.
func (r *analyticsHealthStatusResolver) Influxdb(ctx context.Context, obj *domain.AnalyticsHealthStatus) (string, error) {
	panic("not implemented")
}

// InfluxdbURL is the resolver for the influxdbUrl field.
func (r *analyticsHealthStatusResolver) InfluxdbURL(ctx context.Context, obj *domain.AnalyticsHealthStatus) (string, error) {
	panic("not implemented")
}

// Timestamp is the resolver for the timestamp field.
func (r *analyticsHealthStatusResolver) Timestamp(ctx context.Context, obj *domain.AnalyticsHealthStatus) (*time.Time, error) {
	panic("not implemented")
}

// MicrocontrollerID is the resolver for the microcontrollerId field.
func (r *analyticsReportResolver) MicrocontrollerID(ctx context.Context, obj *domain.AnalyticsReport) (*uuid.UUID, error) {
	panic("not implemented")
}

// Filters is the resolver for the filters field.
func (r *analyticsReportResolver) Filters(ctx context.Context, obj *domain.AnalyticsReport) (*domain.AnalyticsFilter, error) {
	panic("not implemented")
}

// MicrocontrollerID is the resolver for the microcontrollerId field.
func (r *metricResultResolver) MicrocontrollerID(ctx context.Context, obj *domain.MetricResult) (*uuid.UUID, error) {
	panic("not implemented")
}

// Reports is the resolver for the reports field.
func (r *multiMetricReportResolver) Reports(ctx context.Context, obj *domain.MultiReportResponse) ([]*domain.AnalyticsReport, error) {
	panic("not implemented")
}

// GetSingleMetricReport is the resolver for the getSingleMetricReport field.
func (r *queryResolver) GetSingleMetricReport(ctx context.Context, metricName string, controllerID string, filters *generated.AnalyticsFilterInput) (*domain.AnalyticsReport, error) {
	panic("not implemented")
}

// GetMultiMetricReport is the resolver for the getMultiMetricReport field.
func (r *queryResolver) GetMultiMetricReport(ctx context.Context, input domain.MultiMetricReportRequest) (*domain.MultiReportResponse, error) {
	panic("not implemented")
}

// GetTrendAnalysis is the resolver for the getTrendAnalysis field.
func (r *queryResolver) GetTrendAnalysis(ctx context.Context, input domain.TrendAnalysisRequest) (*domain.TrendAnalysis, error) {
	panic("not implemented")
}

// GetSupportedMetrics is the resolver for the getSupportedMetrics field.
func (r *queryResolver) GetSupportedMetrics(ctx context.Context) ([]string, error) {
	panic("not implemented")
}

// GetAnalyticsHealth is the resolver for the getAnalyticsHealth field.
func (r *queryResolver) GetAnalyticsHealth(ctx context.Context) (*domain.AnalyticsHealthStatus, error) {
	panic("not implemented")
}

// MicrocontrollerID is the resolver for the microcontrollerId field.
func (r *trendAnalysisResolver) MicrocontrollerID(ctx context.Context, obj *domain.TrendAnalysis) (*uuid.UUID, error) {
	panic("not implemented")
}

// Filters is the resolver for the filters field.
func (r *trendAnalysisResolver) Filters(ctx context.Context, obj *domain.TrendAnalysis) (*domain.AnalyticsFilter, error) {
	panic("not implemented")
}

// TotalPoints is the resolver for the totalPoints field.
func (r *trendAnalysisResolver) TotalPoints(ctx context.Context, obj *domain.TrendAnalysis) (int, error) {
	panic("not implemented")
}

// AverageValue is the resolver for the averageValue field.
func (r *trendAnalysisResolver) AverageValue(ctx context.Context, obj *domain.TrendAnalysis) (float64, error) {
	panic("not implemented")
}

// MinValue is the resolver for the minValue field.
func (r *trendAnalysisResolver) MinValue(ctx context.Context, obj *domain.TrendAnalysis) (float64, error) {
	panic("not implemented")
}

// MaxValue is the resolver for the maxValue field.
func (r *trendAnalysisResolver) MaxValue(ctx context.Context, obj *domain.TrendAnalysis) (float64, error) {
	panic("not implemented")
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

type analyticsHealthStatusResolver struct{ *Resolver }
type analyticsReportResolver struct{ *Resolver }
type metricResultResolver struct{ *Resolver }
type multiMetricReportResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type trendAnalysisResolver struct{ *Resolver }
