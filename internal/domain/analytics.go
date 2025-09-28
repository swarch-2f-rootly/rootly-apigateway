package domain

import (
	"time"
	"github.com/google/uuid"
)

// TimePeriod represents time intervals for queries
type TimePeriod string

const (
	TimePeriodHour  TimePeriod = "HOUR"
	TimePeriodDay   TimePeriod = "DAY"
	TimePeriodWeek  TimePeriod = "WEEK"
	TimePeriodMonth TimePeriod = "MONTH"
	TimePeriodYear  TimePeriod = "YEAR"
)

// MetricType represents types of metric calculations
type MetricType string

const (
	MetricTypeAverage    MetricType = "AVERAGE"
	MetricTypeMin        MetricType = "MIN"
	MetricTypeMax        MetricType = "MAX"
	MetricTypeMedian     MetricType = "MEDIAN"
	MetricTypePercentile MetricType = "PERCENTILE"
	MetricTypeCount      MetricType = "COUNT"
)

// AnalyticsFilter represents query filters for analytics
type AnalyticsFilter struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     *int       `json:"limit,omitempty"`
}

// MetricResult represents the result of a metric calculation
type MetricResult struct {
	MetricName        string     `json:"metric_name"`
	Value             float64    `json:"value"`
	Unit              string     `json:"unit"`
	CalculatedAt      time.Time  `json:"calculated_at"`
	ControllerID      string     `json:"controller_id"`
	MicrocontrollerID *uuid.UUID `json:"microcontroller_id,omitempty"`
	Description       *string    `json:"description,omitempty"`
}

// AnalyticsReport represents an analytics report for a single controller
type AnalyticsReport struct {
	ControllerID        string           `json:"controller_id"` // Microcontroller uniqueId
	MicrocontrollerID   *uuid.UUID       `json:"microcontroller_id,omitempty"`
	Metrics             []*MetricResult  `json:"metrics"`
	GeneratedAt         time.Time        `json:"generated_at"`
	DataPointsCount     int              `json:"data_points_count"`
	Filters             *AnalyticsFilter `json:"filters,omitempty"`
}

// TrendDataPoint represents a single data point in trend analysis
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Interval  string    `json:"interval"`
}

// TrendAnalysis represents time-series trend analysis
type TrendAnalysis struct {
	MetricName        string             `json:"metric_name"`
	ControllerID      string             `json:"controller_id"`
	MicrocontrollerID *uuid.UUID         `json:"microcontroller_id,omitempty"`
	DataPoints        []*TrendDataPoint  `json:"data_points"`
	Interval          string             `json:"interval"`
	GeneratedAt       time.Time          `json:"generated_at"`
	Filters           *AnalyticsFilter   `json:"filters,omitempty"`
	TotalPoints       int                `json:"total_points"`
	AverageValue      float64            `json:"average_value"`
	MinValue          float64            `json:"min_value"`
	MaxValue          float64            `json:"max_value"`
}

// MultiReportResponse represents analytics for multiple controllers
type MultiReportResponse struct {
	Reports          []*AnalyticsReport `json:"reports"`
	GeneratedAt      time.Time          `json:"generated_at"`
	TotalControllers int                `json:"total_controllers"`
	TotalMetrics     int                `json:"total_metrics"`
}