package repository

import (
	"context"

	"github.com/mikrocloud/mikrocloud/internal/domain/analytics"
)

// MetricRepository handles metric data persistence
type MetricRepository interface {
	Create(ctx context.Context, metric *analytics.Metric) error
	Query(ctx context.Context, query analytics.MetricQuery) ([]*analytics.Metric, error)
	Count(ctx context.Context, query analytics.MetricQuery) (int64, error)
	Aggregate(ctx context.Context, query analytics.MetricQuery) (map[string]float64, error)
	DeleteOlderThan(ctx context.Context, projectID string, cutoff int64) error
}
