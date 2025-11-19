package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/logs"
)

// AnalyticsLogRepository implements LogRepository using the analytics database
type AnalyticsLogRepository struct {
	db *sql.DB
}

// NewAnalyticsLogRepository creates a new analytics-based log repository
func NewAnalyticsLogRepository(db *sql.DB) *AnalyticsLogRepository {
	return &AnalyticsLogRepository{db: db}
}

// Create inserts a new log entry into the analytics database
func (r *AnalyticsLogRepository) Create(ctx context.Context, logEntry *logs.LogEntry) error {
	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(logEntry.Metadata())
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	serviceIDStr := ""
	if logEntry.ServiceID() != nil {
		serviceIDStr = logEntry.ServiceID().String()
	}

	query := `
		INSERT INTO logs (id, project_id, service_id, level, message, source, fields, tags, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// For the analytics database, we'll use fields for metadata and empty tags for now
	emptyTags := "{}"

	_, err = r.db.ExecContext(
		ctx,
		query,
		logEntry.ID().String(),
		logEntry.ProjectID().String(),
		serviceIDStr,
		string(logEntry.Level()),
		logEntry.Message(),
		string(logEntry.Source()),
		string(metadataJSON),
		emptyTags,
		logEntry.Timestamp().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to create log entry: %w", err)
	}

	return nil
}

// Query retrieves log entries based on the provided criteria
func (r *AnalyticsLogRepository) Query(ctx context.Context, query logs.LogQuery) ([]*logs.LogEntry, error) {
	var args []interface{}
	var conditions []string

	sql := `SELECT id, project_id, service_id, level, message, source, fields, timestamp 
			FROM logs`

	// Add WHERE conditions
	conditions = append(conditions, "project_id = ?")
	args = append(args, query.ProjectID.String())

	if query.ServiceID != nil {
		conditions = append(conditions, "service_id = ?")
		args = append(args, query.ServiceID.String())
	}

	if query.Level != nil {
		conditions = append(conditions, "level = ?")
		args = append(args, string(*query.Level))
	}

	if query.Source != nil {
		conditions = append(conditions, "source = ?")
		args = append(args, string(*query.Source))
	}

	if query.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, query.StartTime.Unix())
	}

	if query.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, query.EndTime.Unix())
	}

	if query.Search != "" {
		conditions = append(conditions, "message LIKE ?")
		args = append(args, "%"+query.Search+"%")
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
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logEntries []*logs.LogEntry
	for rows.Next() {
		logEntry, err := r.scanLogEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}
		logEntries = append(logEntries, logEntry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return logEntries, nil
}

// Count returns the number of log entries matching the query
func (r *AnalyticsLogRepository) Count(ctx context.Context, query logs.LogQuery) (int64, error) {
	var args []interface{}
	var conditions []string

	sql := "SELECT COUNT(*) FROM logs"

	// Add WHERE conditions (same logic as Query)
	conditions = append(conditions, "project_id = ?")
	args = append(args, query.ProjectID.String())

	if query.ServiceID != nil {
		conditions = append(conditions, "service_id = ?")
		args = append(args, query.ServiceID.String())
	}

	if query.Level != nil {
		conditions = append(conditions, "level = ?")
		args = append(args, string(*query.Level))
	}

	if query.Source != nil {
		conditions = append(conditions, "source = ?")
		args = append(args, string(*query.Source))
	}

	if query.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, query.StartTime.Unix())
	}

	if query.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, query.EndTime.Unix())
	}

	if query.Search != "" {
		conditions = append(conditions, "message LIKE ?")
		args = append(args, "%"+query.Search+"%")
	}

	if len(conditions) > 0 {
		sql += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count logs: %w", err)
	}

	return count, nil
}

// DeleteOlderThan removes log entries older than the specified cutoff time
func (r *AnalyticsLogRepository) DeleteOlderThan(ctx context.Context, projectID uuid.UUID, cutoff time.Time) error {
	query := "DELETE FROM logs WHERE project_id = ? AND timestamp < ?"

	result, err := r.db.ExecContext(ctx, query, projectID.String(), cutoff.Unix())
	if err != nil {
		return fmt.Errorf("failed to delete old logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	fmt.Printf("Deleted %d old log entries for project %s\n", rowsAffected, projectID.String())
	return nil
}

// scanLogEntry scans a row into a LogEntry object
func (r *AnalyticsLogRepository) scanLogEntry(rows *sql.Rows) (*logs.LogEntry, error) {
	var (
		id, projectIDStr, serviceIDStr, levelStr, message, sourceStr, fieldsJSON string
		timestamp                                                                int64
	)

	err := rows.Scan(
		&id, &projectIDStr, &serviceIDStr, &levelStr,
		&message, &sourceStr, &fieldsJSON, &timestamp,
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

	// Parse metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Create value objects
	logEntryID, err := logs.LogEntryIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid log entry ID: %w", err)
	}

	timestampTime := time.Unix(timestamp, 0)

	// Reconstruct the log entry
	return logs.ReconstructLogEntry(
		logEntryID,
		projectID,
		serviceID,
		logs.LogLevel(levelStr),
		message,
		timestampTime,
		logs.LogSource(sourceStr),
		metadata,
		timestampTime, // Use timestamp as createdAt for now
	), nil
}
