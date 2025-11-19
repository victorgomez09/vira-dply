package analytics_db

import (
	"context"
	"time"
)

// AnalyticsDatabase represents the analytics database interface
type AnalyticsDatabase interface {
	// Core database operations
	Close() error
	Ping(ctx context.Context) error
	DB() interface{} // Access to underlying database connection

	// Analytics operations
	StoreMetric(ctx context.Context, metric Metric) error
	StoreEvent(ctx context.Context, event Event) error
	StoreLogs(ctx context.Context, logs []LogEntry) error

	// Query operations
	GetMetrics(ctx context.Context, filter MetricFilter) ([]Metric, error)
	GetEvents(ctx context.Context, filter EventFilter) ([]Event, error)
	GetLogs(ctx context.Context, filter LogFilter) ([]LogEntry, error)

	// Aggregation operations
	GetMetricAggregation(ctx context.Context, aggregation MetricAggregation) ([]AggregationResult, error)
	GetEventStats(ctx context.Context, timeRange TimeRange) (EventStats, error)
}

// DatabaseType represents the type of analytics database
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	DuckDB     DatabaseType = "duckdb"
	ClickHouse DatabaseType = "clickhouse"
)

// Factory creates an analytics database instance based on configuration
type Factory interface {
	Create(dbType DatabaseType, connectionString string) (AnalyticsDatabase, error)
}

// Data models for analytics

// Metric represents a time-series metric
type Metric struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
	Timestamp time.Time         `json:"timestamp"`
}

// Event represents an application event
type Event struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Source    string            `json:"source"`
	Data      map[string]any    `json:"data"`
	Tags      map[string]string `json:"tags"`
	Timestamp time.Time         `json:"timestamp"`
}

// LogEntry represents a log entry
type LogEntry struct {
	ID        string            `json:"id"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Source    string            `json:"source"`
	Fields    map[string]any    `json:"fields"`
	Tags      map[string]string `json:"tags"`
	Timestamp time.Time         `json:"timestamp"`
}

// Filters for querying

// MetricFilter filters metrics
type MetricFilter struct {
	Names     []string          `json:"names,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	TimeRange TimeRange         `json:"time_range"`
	Limit     int               `json:"limit,omitempty"`
}

// EventFilter filters events
type EventFilter struct {
	Types     []string          `json:"types,omitempty"`
	Sources   []string          `json:"sources,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	TimeRange TimeRange         `json:"time_range"`
	Limit     int               `json:"limit,omitempty"`
}

// LogFilter filters logs
type LogFilter struct {
	Levels    []string          `json:"levels,omitempty"`
	Sources   []string          `json:"sources,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	TimeRange TimeRange         `json:"time_range"`
	Limit     int               `json:"limit,omitempty"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Aggregation types

// MetricAggregation represents metric aggregation configuration
type MetricAggregation struct {
	MetricName string            `json:"metric_name"`
	Function   AggregationFunc   `json:"function"`
	GroupBy    []string          `json:"group_by,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
	TimeRange  TimeRange         `json:"time_range"`
	Interval   time.Duration     `json:"interval,omitempty"`
}

// AggregationFunc represents aggregation functions
type AggregationFunc string

const (
	Sum   AggregationFunc = "sum"
	Avg   AggregationFunc = "avg"
	Min   AggregationFunc = "min"
	Max   AggregationFunc = "max"
	Count AggregationFunc = "count"
)

// AggregationResult represents aggregation result
type AggregationResult struct {
	Value     float64           `json:"value"`
	GroupBy   map[string]string `json:"group_by,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
}

// EventStats represents event statistics
type EventStats struct {
	TotalEvents    int64            `json:"total_events"`
	EventsByType   map[string]int64 `json:"events_by_type"`
	EventsBySource map[string]int64 `json:"events_by_source"`
	TimeRange      TimeRange        `json:"time_range"`
}
