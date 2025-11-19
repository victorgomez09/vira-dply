package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
	"github.com/mikrocloud/mikrocloud/pkg/containers/manager"
)

// StatusSyncService periodically synchronizes database status with container status
type StatusSyncService struct {
	dbService        *DatabaseService
	containerManager manager.ContainerManager
	logger           *slog.Logger
	interval         time.Duration
	stopCh           chan struct{}
}

// NewStatusSyncService creates a new status synchronization service
func NewStatusSyncService(
	dbService *DatabaseService,
	containerManager manager.ContainerManager,
	logger *slog.Logger,
	interval time.Duration,
) *StatusSyncService {
	if interval == 0 {
		interval = 30 * time.Second // Default to 30 seconds
	}

	return &StatusSyncService{
		dbService:        dbService,
		containerManager: containerManager,
		logger:           logger,
		interval:         interval,
		stopCh:           make(chan struct{}),
	}
}

// Start begins the periodic status synchronization
func (s *StatusSyncService) Start(ctx context.Context) {
	s.logger.Info("Starting database status synchronization service", "interval", s.interval)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial sync
	s.syncAllDatabases(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Database status sync service stopped due to context cancellation")
			return
		case <-s.stopCh:
			s.logger.Info("Database status sync service stopped")
			return
		case <-ticker.C:
			s.syncAllDatabases(ctx)
		}
	}
}

// Stop stops the status synchronization service
func (s *StatusSyncService) Stop() {
	close(s.stopCh)
}

// syncAllDatabases retrieves all databases and syncs their status with containers
func (s *StatusSyncService) syncAllDatabases(ctx context.Context) {
	// Get all databases that have container IDs
	databases, err := s.dbService.repo.ListAllWithContainers()
	if err != nil {
		s.logger.Error("Failed to list databases with containers for status sync", "error", err)
		return
	}

	// Sync each database
	var syncedCount, errorCount int
	for _, db := range databases {
		if err := s.SyncDatabaseStatus(ctx, db.ID()); err != nil {
			s.logger.Error("Failed to sync database status",
				"database_id", db.ID().String(),
				"container_id", db.ContainerID(),
				"error", err)
			errorCount++
		} else {
			syncedCount++
		}
	}

	s.logger.Debug("Database status sync completed",
		"total_databases", len(databases),
		"synced", syncedCount,
		"errors", errorCount)
}

// SyncDatabaseStatus synchronizes a specific database status with its container
func (s *StatusSyncService) SyncDatabaseStatus(ctx context.Context, databaseID databases.DatabaseID) error {
	database, err := s.dbService.GetDatabase(ctx, databaseID)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Skip databases without container IDs
	if database.ContainerID() == "" {
		return nil
	}

	// Get container info
	containerInfo, err := s.containerManager.Inspect(ctx, database.ContainerID())
	if err != nil {
		// Container not found - mark database as stopped
		s.logger.Warn("Container not found for database",
			"database_id", database.ID().String(),
			"container_id", database.ContainerID(),
			"error", err)

		if database.Status() != databases.DatabaseStatusStopped {
			return s.dbService.UpdateDatabaseStatus(ctx, databaseID, databases.DatabaseStatusStopped)
		}
		return nil
	}

	// Map container state to database status
	expectedStatus := mapContainerStateToDBStatus(containerInfo.State, containerInfo.Status)

	// Update database status if it differs
	if database.Status() != expectedStatus {
		s.logger.Info("Updating database status based on container state",
			"database_id", database.ID().String(),
			"container_id", database.ContainerID(),
			"container_state", containerInfo.State,
			"container_status", containerInfo.Status,
			"old_status", database.Status(),
			"new_status", expectedStatus)

		return s.dbService.UpdateDatabaseStatus(ctx, databaseID, expectedStatus)
	}

	return nil
}

// mapContainerStateToDBStatus maps container state/status to database status
func mapContainerStateToDBStatus(state, status string) databases.DatabaseStatus {
	state = strings.ToLower(state)
	status = strings.ToLower(status)

	switch state {
	case "running":
		return databases.DatabaseStatusRunning
	case "exited", "stopped":
		return databases.DatabaseStatusStopped
	case "paused":
		return databases.DatabaseStatusStopped
	case "restarting":
		return databases.DatabaseStatusProvisioning
	case "dead", "removing":
		return databases.DatabaseStatusFailed
	default:
		// Check status for more details
		if strings.Contains(status, "up") {
			return databases.DatabaseStatusRunning
		} else if strings.Contains(status, "exited") || strings.Contains(status, "stopped") {
			return databases.DatabaseStatusStopped
		}
		return databases.DatabaseStatusFailed
	}
}

// SyncDatabasesByProject synchronizes all databases in a project
func (s *StatusSyncService) SyncDatabasesByProject(ctx context.Context, projectID string) error {
	// This would require the project ID to be parsed and databases to be listed
	// For now, we'll implement this when we have a GetAllDatabases method
	s.logger.Debug("Project-specific database sync not yet implemented", "project_id", projectID)
	return nil
}
