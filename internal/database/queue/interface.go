package queue_db

import (
	"context"
	"time"
)

// QueueDatabase represents the queue database interface
type QueueDatabase interface {
	// Core database operations
	Close() error
	Ping(ctx context.Context) error

	// Queue operations
	EnqueueTask(ctx context.Context, task Task) error
	EnqueueDelayedTask(ctx context.Context, task Task, delay time.Duration) error
	EnqueueScheduledTask(ctx context.Context, task Task, scheduleTime time.Time) error

	// Task management
	DequeueTask(ctx context.Context, queueName string) (*Task, error)
	RetryTask(ctx context.Context, taskID string, delay time.Duration) error
	ArchiveTask(ctx context.Context, taskID string) error
	DeleteTask(ctx context.Context, taskID string) error

	// Queue inspection
	GetQueueInfo(ctx context.Context, queueName string) (QueueInfo, error)
	GetPendingTasks(ctx context.Context, queueName string, limit int) ([]Task, error)
	GetFailedTasks(ctx context.Context, queueName string, limit int) ([]Task, error)
	GetArchivedTasks(ctx context.Context, queueName string, limit int) ([]Task, error)

	// Queue management
	PurgeQueue(ctx context.Context, queueName string) error
	PauseQueue(ctx context.Context, queueName string) error
	ResumeQueue(ctx context.Context, queueName string) error
}

// DatabaseType represents the type of queue database
type DatabaseType string

const (
	Dragonfly DatabaseType = "dragonfly"
	Redis     DatabaseType = "redis"
)

// Factory creates a queue database instance based on configuration
type Factory interface {
	Create(dbType DatabaseType, connectionString string) (QueueDatabase, error)
}

// Data models for queue system

// Task represents a task in the queue
type Task struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Queue       string            `json:"queue"`
	Payload     []byte            `json:"payload"`
	Options     TaskOptions       `json:"options"`
	Status      TaskStatus        `json:"status"`
	Result      []byte            `json:"result,omitempty"`
	Error       string            `json:"error,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ProcessedAt *time.Time        `json:"processed_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	FailedAt    *time.Time        `json:"failed_at,omitempty"`
	Retries     int               `json:"retries"`
	MaxRetries  int               `json:"max_retries"`
	LastError   string            `json:"last_error,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// TaskOptions represents task configuration options
type TaskOptions struct {
	MaxRetries int           `json:"max_retries"`
	Timeout    time.Duration `json:"timeout"`
	Retention  time.Duration `json:"retention"`
	Unique     bool          `json:"unique"`
	UniqueKey  string        `json:"unique_key,omitempty"`
	Priority   int           `json:"priority"`
	ProcessAt  *time.Time    `json:"process_at,omitempty"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusActive    TaskStatus = "active"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusRetry     TaskStatus = "retry"
	TaskStatusArchived  TaskStatus = "archived"
)

// QueueInfo represents information about a queue
type QueueInfo struct {
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	PendingCount   int64     `json:"pending_count"`
	ActiveCount    int64     `json:"active_count"`
	CompletedCount int64     `json:"completed_count"`
	FailedCount    int64     `json:"failed_count"`
	ArchivedCount  int64     `json:"archived_count"`
	ProcessedTotal int64     `json:"processed_total"`
	IsPaused       bool      `json:"is_paused"`
	LastActivity   time.Time `json:"last_activity"`
	CreatedAt      time.Time `json:"created_at"`
}

// TaskFilter represents filters for querying tasks
type TaskFilter struct {
	Queue       string            `json:"queue,omitempty"`
	Type        string            `json:"type,omitempty"`
	Status      TaskStatus        `json:"status,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	CreatedFrom *time.Time        `json:"created_from,omitempty"`
	CreatedTo   *time.Time        `json:"created_to,omitempty"`
	Limit       int               `json:"limit,omitempty"`
	Offset      int               `json:"offset,omitempty"`
}

// QueueStats represents statistics for queues
type QueueStats struct {
	TotalQueues    int                  `json:"total_queues"`
	TotalTasks     int64                `json:"total_tasks"`
	QueueStats     map[string]QueueInfo `json:"queue_stats"`
	StatusCounts   map[TaskStatus]int64 `json:"status_counts"`
	TypeCounts     map[string]int64     `json:"type_counts"`
	ProcessingRate float64              `json:"processing_rate"` // tasks per minute
	FailureRate    float64              `json:"failure_rate"`    // percentage
}
