package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/analytics"
	"github.com/mikrocloud/mikrocloud/internal/domain/analytics/repository"
)

// AnalyticsService handles analytics business logic
type AnalyticsService struct {
	metricRepo repository.MetricRepository
}

func NewAnalyticsService(metricRepo repository.MetricRepository) *AnalyticsService {
	return &AnalyticsService{
		metricRepo: metricRepo,
	}
}

// RecordMetric records a new metric
func (s *AnalyticsService) RecordMetric(
	ctx context.Context,
	projectID string,
	serviceID *string,
	name string,
	value float64,
	unit string,
	tags map[string]string,
) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	// Convert projectID to UUID
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Convert serviceID to UUID if provided
	var serviceUUID *uuid.UUID
	if serviceID != nil {
		parsed, err := uuid.Parse(*serviceID)
		if err != nil {
			return fmt.Errorf("invalid service ID: %w", err)
		}
		serviceUUID = &parsed
	}

	// Create metric name value object
	metricName, err := analytics.NewMetricName(name)
	if err != nil {
		return fmt.Errorf("invalid metric name: %w", err)
	}

	// Create metric
	metric := analytics.NewMetric(
		projectUUID,
		serviceUUID,
		metricName,
		value,
		analytics.MetricUnit(unit),
		tags,
		time.Now(),
	)

	return s.metricRepo.Create(ctx, metric)
}

// QueryMetrics queries metrics based on criteria
func (s *AnalyticsService) QueryMetrics(ctx context.Context, query analytics.MetricQuery) ([]*analytics.Metric, error) {
	return s.metricRepo.Query(ctx, query)
}

// CountMetrics counts metrics matching the query
func (s *AnalyticsService) CountMetrics(ctx context.Context, query analytics.MetricQuery) (int64, error) {
	return s.metricRepo.Count(ctx, query)
}

// AggregateMetrics aggregates metrics
func (s *AnalyticsService) AggregateMetrics(ctx context.Context, query analytics.MetricQuery) (map[string]float64, error) {
	return s.metricRepo.Aggregate(ctx, query)
}

// CleanupOldMetrics removes metrics older than the specified timestamp
func (s *AnalyticsService) CleanupOldMetrics(ctx context.Context, projectID string, cutoffDays int) error {
	if cutoffDays <= 0 {
		return fmt.Errorf("cutoff days must be positive")
	}

	cutoff := int64(cutoffDays)
	return s.metricRepo.DeleteOlderThan(ctx, projectID, cutoff)
}
