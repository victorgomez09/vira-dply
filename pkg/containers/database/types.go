package database

import (
	"context"

	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
)

// DatabaseDeploymentService handles database container deployment and management
type DatabaseDeploymentService interface {
	// Deploy creates and starts a database container
	Deploy(ctx context.Context, database *databases.Database) (*DeploymentResult, error)

	// Start starts an existing database container
	Start(ctx context.Context, database *databases.Database) error

	// Stop stops a running database container
	Stop(ctx context.Context, database *databases.Database) error

	// Remove removes a database container completely
	Remove(ctx context.Context, database *databases.Database) error

	// Restart recreates and restarts a database container with updated configuration
	Restart(ctx context.Context, database *databases.Database) error

	// GetStatus returns the current status of a database container
	GetStatus(ctx context.Context, database *databases.Database) (*ContainerStatus, error)

	// GetLogs retrieves logs from a database container
	GetLogs(ctx context.Context, database *databases.Database, follow bool) ([]byte, error)
}

// DeploymentResult contains information about a deployed database container
type DeploymentResult struct {
	ContainerID   string
	ContainerName string
	Port          string
	Status        string
	CreatedAt     string
}

// ContainerStatus represents the current state of a database container
type ContainerStatus struct {
	ID        string
	Name      string
	State     string
	Status    string
	Port      string
	Health    string
	StartedAt string
}

// DatabaseContainerConfig represents the configuration for deploying a database container
type DatabaseContainerConfig struct {
	Database      *databases.Database
	Image         string
	ContainerName string
	Port          string
	Environment   map[string]string
	Volumes       map[string]string
	Command       []string
	HealthCheck   *HealthCheckConfig
}

// HealthCheckConfig defines health check parameters for database containers
type HealthCheckConfig struct {
	Test     []string
	Interval string
	Timeout  string
	Retries  int
}

// DatabaseImageResolver resolves the appropriate Docker image for each database type
type DatabaseImageResolver interface {
	ResolveImage(dbType databases.DatabaseType, version string) string
	GetDefaultVersion(dbType databases.DatabaseType) string
	GetSupportedVersions(dbType databases.DatabaseType) []string
}

// ContainerConfigBuilder builds container configuration for different database types
type ContainerConfigBuilder interface {
	BuildConfig(database *databases.Database) (*DatabaseContainerConfig, error)
	BuildPostgreSQLConfig(database *databases.Database) (*DatabaseContainerConfig, error)
	BuildMySQLConfig(database *databases.Database) (*DatabaseContainerConfig, error)
	BuildMariaDBConfig(database *databases.Database) (*DatabaseContainerConfig, error)
	BuildRedisConfig(database *databases.Database) (*DatabaseContainerConfig, error)
	BuildKeyDBConfig(database *databases.Database) (*DatabaseContainerConfig, error)
	BuildDragonflyConfig(database *databases.Database) (*DatabaseContainerConfig, error)
	BuildMongoDBConfig(database *databases.Database) (*DatabaseContainerConfig, error)
	BuildClickHouseConfig(database *databases.Database) (*DatabaseContainerConfig, error)
}
