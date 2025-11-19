package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/logs"
	"github.com/mikrocloud/mikrocloud/internal/domain/logs/repository"
)

type LogService struct {
	logRepo repository.LogRepository
}

func NewLogService(logRepo repository.LogRepository) *LogService {
	return &LogService{
		logRepo: logRepo,
	}
}

func (s *LogService) CreateLogEntry(
	ctx context.Context,
	projectID uuid.UUID,
	serviceID *uuid.UUID,
	level logs.LogLevel,
	message string,
	source logs.LogSource,
	metadata map[string]any,
) error {
	if message == "" {
		return fmt.Errorf("log message cannot be empty")
	}

	logEntry := logs.NewLogEntry(projectID, serviceID, level, message, source, metadata)
	return s.logRepo.Create(ctx, logEntry)
}

func (s *LogService) QueryLogs(ctx context.Context, query logs.LogQuery) ([]*logs.LogEntry, error) {
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 1000 {
		query.Limit = 1000
	}

	return s.logRepo.Query(ctx, query)
}

func (s *LogService) CountLogs(ctx context.Context, query logs.LogQuery) (int64, error) {
	return s.logRepo.Count(ctx, query)
}

func (s *LogService) CleanupOldLogs(ctx context.Context, projectID uuid.UUID, retentionDays int) error {
	if retentionDays <= 0 {
		return fmt.Errorf("retention days must be positive")
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return s.logRepo.DeleteOlderThan(ctx, projectID, cutoff)
}
