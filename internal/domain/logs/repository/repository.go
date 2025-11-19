package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/logs"
)

type LogRepository interface {
	Create(ctx context.Context, logEntry *logs.LogEntry) error
	Query(ctx context.Context, query logs.LogQuery) ([]*logs.LogEntry, error)
	Count(ctx context.Context, query logs.LogQuery) (int64, error)
	DeleteOlderThan(ctx context.Context, projectID uuid.UUID, cutoff time.Time) error
}
