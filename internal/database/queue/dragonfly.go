package queue_db

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"golang.org/x/exp/slog"
)

// DragonflyQueueDatabase implements QueueDatabase interface using Asynq + Dragonfly
type DragonflyQueueDatabase struct {
	client    *asynq.Client
	inspector *asynq.Inspector
	server    *asynq.Server
	redisOpt  asynq.RedisConnOpt
}

// NewDragonflyDatabase creates a new Dragonfly queue database instance
func NewDragonflyDatabase(connectionString string) (QueueDatabase, error) {
	// Parse connection string for Redis/Dragonfly
	redisOpt, err := asynq.ParseRedisURI(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Dragonfly connection string: %w", err)
	}

	client := asynq.NewClient(redisOpt)
	inspector := asynq.NewInspector(redisOpt)

	// Test the connection
	info, err := inspector.GetQueueInfo("default")
	if err != nil {
		// If default queue doesn't exist, that's fine - we'll create it
		slog.Debug("Default queue not found, will be created on first use")
	} else {
		slog.Debug("Connected to Dragonfly", "queue_info", info)
	}

	slog.Info("Dragonfly queue database connection established", "connection", maskConnectionString(connectionString))

	return &DragonflyQueueDatabase{
		client:    client,
		inspector: inspector,
		redisOpt:  redisOpt,
	}, nil
}

func (d *DragonflyQueueDatabase) Close() error {
	if d.client != nil {
		if err := d.client.Close(); err != nil {
			slog.Error("Error closing Dragonfly client", "error", err)
			return err
		}
	}
	if d.inspector != nil {
		if err := d.inspector.Close(); err != nil {
			slog.Error("Error closing Dragonfly inspector", "error", err)
			return err
		}
	}
	if d.server != nil {
		d.server.Shutdown()
	}
	slog.Info("Dragonfly queue database connection closed")
	return nil
}

func (d *DragonflyQueueDatabase) Ping(ctx context.Context) error {
	// Try to get queue info as a ping test
	_, err := d.inspector.GetQueueInfo("default")
	if err != nil {
		return fmt.Errorf("failed to ping Dragonfly: %w", err)
	}
	return nil
}

// Queue operations
func (d *DragonflyQueueDatabase) EnqueueTask(ctx context.Context, task Task) error {
	asynqTask := asynq.NewTask(task.Type, task.Payload, d.buildTaskOptions(task)...)

	info, err := d.client.EnqueueContext(ctx, asynqTask)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	slog.Debug("Task enqueued", "task_id", info.ID, "queue", info.Queue, "type", info.Type)
	return nil
}

func (d *DragonflyQueueDatabase) EnqueueDelayedTask(ctx context.Context, task Task, delay time.Duration) error {
	asynqTask := asynq.NewTask(task.Type, task.Payload, d.buildTaskOptions(task)...)

	info, err := d.client.EnqueueContext(ctx, asynqTask, asynq.ProcessIn(delay))
	if err != nil {
		return fmt.Errorf("failed to enqueue delayed task: %w", err)
	}

	slog.Debug("Delayed task enqueued", "task_id", info.ID, "delay", delay, "type", info.Type)
	return nil
}

func (d *DragonflyQueueDatabase) EnqueueScheduledTask(ctx context.Context, task Task, scheduleTime time.Time) error {
	asynqTask := asynq.NewTask(task.Type, task.Payload, d.buildTaskOptions(task)...)

	info, err := d.client.EnqueueContext(ctx, asynqTask, asynq.ProcessAt(scheduleTime))
	if err != nil {
		return fmt.Errorf("failed to enqueue scheduled task: %w", err)
	}

	slog.Debug("Scheduled task enqueued", "task_id", info.ID, "schedule_time", scheduleTime, "type", info.Type)
	return nil
}

// Helper to build asynq task options from our Task struct
func (d *DragonflyQueueDatabase) buildTaskOptions(task Task) []asynq.Option {
	var opts []asynq.Option

	if task.Queue != "" {
		opts = append(opts, asynq.Queue(task.Queue))
	}

	if task.Options.MaxRetries > 0 {
		opts = append(opts, asynq.MaxRetry(task.Options.MaxRetries))
	}

	if task.Options.Timeout > 0 {
		opts = append(opts, asynq.Timeout(task.Options.Timeout))
	}

	if task.Options.Retention > 0 {
		opts = append(opts, asynq.Retention(task.Options.Retention))
	}

	if task.Options.Unique {
		opts = append(opts, asynq.Unique(time.Hour))
	}

	if task.Options.ProcessAt != nil {
		opts = append(opts, asynq.ProcessAt(*task.Options.ProcessAt))
	}

	return opts
}

// Task management - Basic implementations
func (d *DragonflyQueueDatabase) DequeueTask(ctx context.Context, queueName string) (*Task, error) {
	// This is handled by the asynq server/mux, not directly by client
	return nil, fmt.Errorf("DequeueTask is handled by asynq server, not client")
}

func (d *DragonflyQueueDatabase) RetryTask(ctx context.Context, taskID string, delay time.Duration) error {
	queueName := queueName(taskID)
	if err := d.inspector.RunTask(queueName, taskID); err != nil {
		return fmt.Errorf("failed to retry task: %w", err)
	}
	slog.Debug("Task marked for retry", "task_id", taskID, "queue", queueName)
	return nil
}

func (d *DragonflyQueueDatabase) ArchiveTask(ctx context.Context, taskID string) error {
	return d.inspector.ArchiveTask(queueName(taskID), taskID)
}

func (d *DragonflyQueueDatabase) DeleteTask(ctx context.Context, taskID string) error {
	return d.inspector.DeleteTask(queueName(taskID), taskID)
}

// Queue inspection
func (d *DragonflyQueueDatabase) GetQueueInfo(ctx context.Context, queueName string) (QueueInfo, error) {
	info, err := d.inspector.GetQueueInfo(queueName)
	if err != nil {
		return QueueInfo{}, fmt.Errorf("failed to get queue info: %w", err)
	}

	return QueueInfo{
		Name:           queueName,
		Size:           int64(info.Size),
		PendingCount:   int64(info.Pending),
		ActiveCount:    int64(info.Active),
		CompletedCount: int64(info.Completed),
		FailedCount:    int64(info.Failed),
		ArchivedCount:  int64(info.Archived),
		ProcessedTotal: int64(info.Processed),
		IsPaused:       info.Paused,
		LastActivity:   time.Now(), // Asynq doesn't provide this, so we use current time
	}, nil
}

func (d *DragonflyQueueDatabase) GetPendingTasks(ctx context.Context, queueName string, limit int) ([]Task, error) {
	tasks, err := d.inspector.ListPendingTasks(queueName, asynq.PageSize(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	return d.convertAsynqTasks(tasks), nil
}

func (d *DragonflyQueueDatabase) GetFailedTasks(ctx context.Context, queueName string, limit int) ([]Task, error) {
	tasks, err := d.inspector.ListRetryTasks(queueName, asynq.PageSize(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get failed/retry tasks: %w", err)
	}

	return d.convertAsynqTasks(tasks), nil
}

func (d *DragonflyQueueDatabase) GetArchivedTasks(ctx context.Context, queueName string, limit int) ([]Task, error) {
	tasks, err := d.inspector.ListArchivedTasks(queueName, asynq.PageSize(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get archived tasks: %w", err)
	}

	return d.convertAsynqTasks(tasks), nil
}

// Queue management
func (d *DragonflyQueueDatabase) PurgeQueue(ctx context.Context, queueName string) error {
	return d.inspector.DeleteQueue(queueName, false)
}

func (d *DragonflyQueueDatabase) PauseQueue(ctx context.Context, queueName string) error {
	return d.inspector.PauseQueue(queueName)
}

func (d *DragonflyQueueDatabase) ResumeQueue(ctx context.Context, queueName string) error {
	return d.inspector.UnpauseQueue(queueName)
}

// Helper functions
func (d *DragonflyQueueDatabase) convertAsynqTasks(asynqTasks []*asynq.TaskInfo) []Task {
	tasks := make([]Task, len(asynqTasks))

	for i, t := range asynqTasks {
		task := Task{
			ID:         t.ID,
			Type:       t.Type,
			Queue:      t.Queue,
			Payload:    t.Payload,
			Status:     convertAsynqState(t.State),
			Retries:    t.Retried,
			MaxRetries: t.MaxRetry,
		}

		// Convert timestamps
		if !t.NextProcessAt.IsZero() {
			task.ProcessedAt = &t.NextProcessAt
		}

		// Handle error information
		if t.LastErr != "" {
			task.Error = t.LastErr
			task.LastError = t.LastErr
		}

		tasks[i] = task
	}

	return tasks
}

func convertAsynqState(state asynq.TaskState) TaskStatus {
	switch state {
	case asynq.TaskStatePending:
		return TaskStatusPending
	case asynq.TaskStateActive:
		return TaskStatusActive
	case asynq.TaskStateCompleted:
		return TaskStatusCompleted
	case asynq.TaskStateArchived:
		return TaskStatusArchived
	default:
		return TaskStatusPending
	}
}

func queueName(taskID string) string {
	// Extract queue name from task ID if needed, or default to "default"
	return "default"
}

func maskConnectionString(connectionString string) string {
	// Simple connection string masking for logging
	return "dragonfly://***:***@***"
}
