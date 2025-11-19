package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/analytics"
)

// SQLiteMetricRepository implements MetricRepository using SQLite
type SQLiteMetricRepository struct {
	db *sql.DB
}

// NewSQLiteMetricRepository creates a new SQLite-based metric repository
func NewSQLiteMetricRepository(db *sql.DB) *SQLiteMetricRepository {
	return &SQLiteMetricRepository{db: db}
}

// Create inserts a new metric into the database
func (r *SQLiteMetricRepository) Create(ctx context.Context, metric *analytics.Metric) error {
	tagsJSON, err := json.Marshal(metric.Tags())
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	serviceIDStr := ""
	if metric.ServiceID() != nil {
		serviceIDStr = metric.ServiceID().String()
	}

	query := `
		INSERT INTO metrics (id, project_id, service_id, name, value, unit, tags, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		metric.ID().String(),
		metric.ProjectID().String(),
		serviceIDStr,
		metric.Name().String(),
		metric.Value(),
		string(metric.Unit()),
		string(tagsJSON),
		metric.Timestamp().Unix(),
		metric.CreatedAt().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to create metric: %w", err)
	}

	return nil
}

// Query retrieves metrics based on the provided criteria
func (r *SQLiteMetricRepository) Query(ctx context.Context, query analytics.MetricQuery) ([]*analytics.Metric, error) {
	var args []interface{}
	var conditions []string

	sql := "SELECT id, project_id, service_id, name, value, unit, tags, timestamp, created_at FROM metrics"

	// Add WHERE conditions
	conditions = append(conditions, "project_id = ?")
	args = append(args, query.ProjectID.String())

	if query.ServiceID != nil {
		conditions = append(conditions, "service_id = ?")
		args = append(args, query.ServiceID.String())
	}

	if query.Name != nil {
		conditions = append(conditions, "name = ?")
		args = append(args, query.Name.String())
	}

	if query.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, query.StartTime.Unix())
	}

	if query.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, query.EndTime.Unix())
	}

	// Add WHERE clause if conditions exist
	if len(conditions) > 0 {
		sql += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY
	sql += " ORDER BY timestamp DESC"

	// Add LIMIT and OFFSET
	if query.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", query.Limit)
	}
	if query.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", query.Offset)
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.Metric
	for rows.Next() {
		metric, err := r.scanMetric(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return metrics, nil
}

// Count returns the number of metrics matching the query
func (r *SQLiteMetricRepository) Count(ctx context.Context, query analytics.MetricQuery) (int64, error) {
	var args []interface{}
	var conditions []string

	sql := "SELECT COUNT(*) FROM metrics"

	// Add WHERE conditions (same logic as Query)
	conditions = append(conditions, "project_id = ?")
	args = append(args, query.ProjectID.String())

	if query.ServiceID != nil {
		conditions = append(conditions, "service_id = ?")
		args = append(args, query.ServiceID.String())
	}

	if query.Name != nil {
		conditions = append(conditions, "name = ?")
		args = append(args, query.Name.String())
	}

	if query.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, query.StartTime.Unix())
	}

	if query.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, query.EndTime.Unix())
	}

	if len(conditions) > 0 {
		sql += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count metrics: %w", err)
	}

	return count, nil
}

// Aggregate performs aggregation on metrics
func (r *SQLiteMetricRepository) Aggregate(ctx context.Context, query analytics.MetricQuery) (map[string]float64, error) {
	var args []interface{}
	var conditions []string

	sql := `
		SELECT 
			name,
			AVG(value) as avg_value,
			MIN(value) as min_value,
			MAX(value) as max_value,
			SUM(value) as sum_value,
			COUNT(*) as count_value
		FROM metrics
	`

	// Add WHERE conditions (same logic as Query)
	conditions = append(conditions, "project_id = ?")
	args = append(args, query.ProjectID.String())

	if query.ServiceID != nil {
		conditions = append(conditions, "service_id = ?")
		args = append(args, query.ServiceID.String())
	}

	if query.Name != nil {
		conditions = append(conditions, "name = ?")
		args = append(args, query.Name.String())
	}

	if query.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, query.StartTime.Unix())
	}

	if query.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, query.EndTime.Unix())
	}

	if len(conditions) > 0 {
		sql += " WHERE " + strings.Join(conditions, " AND ")
	}

	sql += " GROUP BY name"

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate metrics: %w", err)
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var name string
		var avg, min, max, sum, count float64

		err := rows.Scan(&name, &avg, &min, &max, &sum, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan aggregate row: %w", err)
		}

		result[name+"_avg"] = avg
		result[name+"_min"] = min
		result[name+"_max"] = max
		result[name+"_sum"] = sum
		result[name+"_count"] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}

// DeleteOlderThan removes metrics older than the specified cutoff (in days)
func (r *SQLiteMetricRepository) DeleteOlderThan(ctx context.Context, projectID string, cutoffDays int64) error {
	cutoffTime := time.Now().AddDate(0, 0, -int(cutoffDays)).Unix()

	query := "DELETE FROM metrics WHERE project_id = ? AND timestamp < ?"

	result, err := r.db.ExecContext(ctx, query, projectID, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to delete old metrics: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	fmt.Printf("Deleted %d old metrics for project %s\n", rowsAffected, projectID)
	return nil
}

// scanMetric scans a row into a Metric object
func (r *SQLiteMetricRepository) scanMetric(rows *sql.Rows) (*analytics.Metric, error) {
	var (
		id, projectIDStr, serviceIDStr, nameStr, unitStr, tagsJSON string
		value                                                      float64
		timestamp, createdAt                                       int64
	)

	err := rows.Scan(
		&id, &projectIDStr, &serviceIDStr, &nameStr,
		&value, &unitStr, &tagsJSON, &timestamp, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse UUIDs
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	var serviceID *uuid.UUID
	if serviceIDStr != "" {
		parsed, err := uuid.Parse(serviceIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid service ID: %w", err)
		}
		serviceID = &parsed
	}

	// Parse tags
	var tags map[string]string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	// Create value objects
	metricID := analytics.MetricIDFromString(id)

	metricName, err := analytics.NewMetricName(nameStr)
	if err != nil {
		return nil, fmt.Errorf("invalid metric name: %w", err)
	}

	// Reconstruct the metric
	return analytics.ReconstructMetric(
		metricID,
		projectID,
		serviceID,
		metricName,
		value,
		analytics.MetricUnit(unitStr),
		tags,
		time.Unix(timestamp, 0),
		time.Unix(createdAt, 0),
	), nil
}
