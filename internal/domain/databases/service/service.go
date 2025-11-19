package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
	"github.com/mikrocloud/mikrocloud/internal/domain/disks"
	"github.com/mikrocloud/mikrocloud/pkg/containers/database"
)

type DatabaseRepository interface {
	Create(database *databases.Database) error
	GetByID(id databases.DatabaseID) (*databases.Database, error)
	GetByName(projectID uuid.UUID, name databases.DatabaseName) (*databases.Database, error)
	ListByProject(projectID uuid.UUID) ([]*databases.Database, error)
	ListByEnvironment(projectID, environmentID uuid.UUID) ([]*databases.Database, error)
	ListAllWithContainers() ([]*databases.Database, error)
	Update(database *databases.Database) error
	Delete(id databases.DatabaseID) error
	ExistsByName(projectID uuid.UUID, name databases.DatabaseName) (bool, error)
}

type DiskService interface {
	CreateDisk(ctx context.Context, name disks.DiskName, projectID uuid.UUID, size disks.DiskSize, mountPath string, filesystem disks.Filesystem, persistent bool) (*disks.Disk, error)
	AttachDisk(ctx context.Context, diskID disks.DiskID, serviceID uuid.UUID) error
}

type DatabaseService struct {
	repo                DatabaseRepository
	containerDeployment database.DatabaseDeploymentService
	diskService         DiskService
}

func NewDatabaseService(repo DatabaseRepository, containerDeployment database.DatabaseDeploymentService, diskService DiskService) *DatabaseService {
	return &DatabaseService{
		repo:                repo,
		containerDeployment: containerDeployment,
		diskService:         diskService,
	}
}

type CreateDatabaseCommand struct {
	Name          string
	Description   string
	Type          databases.DatabaseType
	ProjectID     uuid.UUID
	EnvironmentID uuid.UUID
	Config        *databases.DatabaseConfig
}

func (s *DatabaseService) CreateDatabase(ctx context.Context, cmd CreateDatabaseCommand) (*databases.Database, error) {
	name, err := databases.NewDatabaseName(cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid database name: %w", err)
	}

	// Check if database already exists
	exists, err := s.repo.ExistsByName(cmd.ProjectID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("database with name %s already exists in project", name.String())
	}

	// Generate default config if not provided
	var config databases.DatabaseConfig
	if cmd.Config != nil {
		config = *cmd.Config
	} else {
		config = s.generateDefaultConfig(cmd.Type)
	}

	database := databases.NewDatabase(
		name,
		cmd.Description,
		cmd.Type,
		cmd.ProjectID,
		cmd.EnvironmentID,
		config,
	)

	// Generate connection string
	connectionString := database.GenerateConnectionString()
	database.SetConnectionString(connectionString)

	// Set default ports
	ports := map[string]int{
		"main": database.GetMainPort(),
	}

	// Add additional ports for specific database types
	switch cmd.Type {
	case databases.DatabaseTypeClickHouse:
		if config.ClickHouse != nil {
			ports["http"] = config.ClickHouse.HTTPPort
		}
	}

	database.SetPorts(ports)

	if err := s.repo.Create(database); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return database, nil
}

func (s *DatabaseService) GetDatabase(ctx context.Context, id databases.DatabaseID) (*databases.Database, error) {
	database, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}
	return database, nil
}

func (s *DatabaseService) GetDatabaseByName(ctx context.Context, projectID uuid.UUID, name string) (*databases.Database, error) {
	dbName, err := databases.NewDatabaseName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid database name: %w", err)
	}

	database, err := s.repo.GetByName(projectID, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database by name: %w", err)
	}
	return database, nil
}

type UpdateDatabaseCommand struct {
	ID          databases.DatabaseID
	Description *string
	Config      *databases.DatabaseConfig
}

func (s *DatabaseService) UpdateDatabase(ctx context.Context, cmd UpdateDatabaseCommand) (*databases.Database, error) {
	database, err := s.repo.GetByID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("database not found: %w", err)
	}

	if cmd.Description != nil {
		database.UpdateDescription(*cmd.Description)
	}

	if cmd.Config != nil {
		database.UpdateConfig(*cmd.Config)

		// Regenerate connection string if config changed
		connectionString := database.GenerateConnectionString()
		database.SetConnectionString(connectionString)

		// Update ports if needed
		ports := map[string]int{
			"main": database.GetMainPort(),
		}

		// Add additional ports for specific database types
		switch database.Type() {
		case databases.DatabaseTypeClickHouse:
			if cmd.Config.ClickHouse != nil {
				ports["http"] = cmd.Config.ClickHouse.HTTPPort
			}
		}

		database.SetPorts(ports)
	}

	if err := s.repo.Update(database); err != nil {
		return nil, fmt.Errorf("failed to update database: %w", err)
	}

	return database, nil
}

func (s *DatabaseService) DeleteDatabase(ctx context.Context, id databases.DatabaseID) error {
	// Check if database exists and validate deletion
	database, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("database not found: %w", err)
	}

	if err := database.CanDelete(); err != nil {
		return fmt.Errorf("cannot delete database: %w", err)
	}

	// Mark as deleting
	database.ChangeStatus(databases.DatabaseStatusDeleting)
	if err := s.repo.Update(database); err != nil {
		return fmt.Errorf("failed to update database status: %w", err)
	}

	// Remove the container if it exists
	if database.ContainerID() != "" {
		if err := s.containerDeployment.Remove(ctx, database); err != nil {
			// Log the error but continue with database deletion
			// This ensures we can clean up the database record even if container removal fails
			fmt.Printf("Warning: failed to remove database container: %v\n", err)
		}
	}

	// Delete the database record
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	return nil
}

func (s *DatabaseService) ListDatabases(ctx context.Context, projectID uuid.UUID) ([]*databases.Database, error) {
	databases, err := s.repo.ListByProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	return databases, nil
}

func (s *DatabaseService) ListDatabasesByEnvironment(ctx context.Context, projectID, environmentID uuid.UUID) ([]*databases.Database, error) {
	databases, err := s.repo.ListByEnvironment(projectID, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases by environment: %w", err)
	}
	return databases, nil
}

func (s *DatabaseService) StartDatabase(ctx context.Context, id databases.DatabaseID) error {
	database, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("database not found: %w", err)
	}

	if err := database.CanStart(); err != nil {
		return fmt.Errorf("cannot start database: %w", err)
	}

	database.ChangeStatus(databases.DatabaseStatusProvisioning)

	if err := s.repo.Update(database); err != nil {
		return fmt.Errorf("failed to update database status: %w", err)
	}

	// Create and attach default disk for persistent storage databases if disk service is available
	if s.diskService != nil && s.requiresPersistentStorage(database.Type()) {
		if err := s.createDefaultDisk(ctx, database); err != nil {
			// Log error but don't fail the deployment
			fmt.Printf("Warning: failed to create default disk for database %s: %v\n", database.ID(), err)
		}
	}

	// Deploy the database container
	deployResult, err := s.containerDeployment.Deploy(ctx, database)
	if err != nil {
		// Mark as failed if deployment fails
		database.ChangeStatus(databases.DatabaseStatusFailed)
		_ = s.repo.Update(database)
		return fmt.Errorf("failed to deploy database container: %w", err)
	}

	// Update database with container information
	database.ChangeStatus(databases.DatabaseStatusRunning)
	database.SetContainerID(deployResult.ContainerID)

	if err := s.repo.Update(database); err != nil {
		return fmt.Errorf("failed to update database status: %w", err)
	}

	return nil
}

func (s *DatabaseService) StopDatabase(ctx context.Context, id databases.DatabaseID) error {
	database, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("database not found: %w", err)
	}

	if err := database.CanStop(); err != nil {
		return fmt.Errorf("cannot stop database: %w", err)
	}

	// Stop the container if it exists
	if database.ContainerID() != "" {
		if err := s.containerDeployment.Stop(ctx, database); err != nil {
			return fmt.Errorf("failed to stop database container: %w", err)
		}
	}

	database.ChangeStatus(databases.DatabaseStatusStopped)

	if err := s.repo.Update(database); err != nil {
		return fmt.Errorf("failed to update database status: %w", err)
	}

	return nil
}

func (s *DatabaseService) UpdateDatabaseStatus(ctx context.Context, id databases.DatabaseID, status databases.DatabaseStatus) error {
	database, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("database not found: %w", err)
	}

	database.ChangeStatus(status)

	if err := s.repo.Update(database); err != nil {
		return fmt.Errorf("failed to update database status: %w", err)
	}

	return nil
}

// generateDefaultConfig creates a default configuration for the specified database type
func (s *DatabaseService) generateDefaultConfig(dbType databases.DatabaseType) databases.DatabaseConfig {
	config := databases.DatabaseConfig{Type: dbType}

	switch dbType {
	case databases.DatabaseTypePostgreSQL:
		config.PostgreSQL = databases.DefaultPostgreSQLConfig()
	case databases.DatabaseTypeMySQL:
		config.MySQL = databases.DefaultMySQLConfig()
	case databases.DatabaseTypeMariaDB:
		config.MariaDB = databases.DefaultMariaDBConfig()
	case databases.DatabaseTypeRedis:
		config.Redis = databases.DefaultRedisConfig()
	case databases.DatabaseTypeKeyDB:
		config.KeyDB = databases.DefaultKeyDBConfig()
	case databases.DatabaseTypeDragonfly:
		config.Dragonfly = databases.DefaultDragonflyConfig()
	case databases.DatabaseTypeMongoDB:
		config.MongoDB = databases.DefaultMongoDBConfig()
	case databases.DatabaseTypeClickHouse:
		config.ClickHouse = databases.DefaultClickHouseConfig()
	}

	return config
}

// GetSupportedDatabaseTypes returns all supported database types
func (s *DatabaseService) GetSupportedDatabaseTypes() []databases.DatabaseType {
	return []databases.DatabaseType{
		databases.DatabaseTypePostgreSQL,
		databases.DatabaseTypeMySQL,
		databases.DatabaseTypeMariaDB,
		databases.DatabaseTypeRedis,
		databases.DatabaseTypeKeyDB,
		databases.DatabaseTypeDragonfly,
		databases.DatabaseTypeMongoDB,
		databases.DatabaseTypeClickHouse,
	}
}

// ValidateDatabaseConfig validates a database configuration
func (s *DatabaseService) ValidateDatabaseConfig(dbType databases.DatabaseType, config databases.DatabaseConfig) error {
	if config.Type != dbType {
		return fmt.Errorf("config type %s does not match database type %s", config.Type, dbType)
	}

	switch dbType {
	case databases.DatabaseTypePostgreSQL:
		if config.PostgreSQL == nil {
			return fmt.Errorf("PostgreSQL configuration is required")
		}
		if config.PostgreSQL.Username == "" {
			return fmt.Errorf("PostgreSQL username is required")
		}
		if config.PostgreSQL.DatabaseName == "" {
			return fmt.Errorf("PostgreSQL database name is required")
		}
	case databases.DatabaseTypeMySQL:
		if config.MySQL == nil {
			return fmt.Errorf("MySQL configuration is required")
		}
		if config.MySQL.Username == "" {
			return fmt.Errorf("MySQL username is required")
		}
		if config.MySQL.DatabaseName == "" {
			return fmt.Errorf("MySQL database name is required")
		}
	case databases.DatabaseTypeMariaDB:
		if config.MariaDB == nil {
			return fmt.Errorf("MariaDB configuration is required")
		}
		if config.MariaDB.Username == "" {
			return fmt.Errorf("MariaDB username is required")
		}
		if config.MariaDB.DatabaseName == "" {
			return fmt.Errorf("MariaDB database name is required")
		}
	case databases.DatabaseTypeRedis:
		if config.Redis == nil {
			return fmt.Errorf("Redis configuration is required")
		}
	case databases.DatabaseTypeKeyDB:
		if config.KeyDB == nil {
			return fmt.Errorf("KeyDB configuration is required")
		}
	case databases.DatabaseTypeDragonfly:
		if config.Dragonfly == nil {
			return fmt.Errorf("Dragonfly configuration is required")
		}
	case databases.DatabaseTypeMongoDB:
		if config.MongoDB == nil {
			return fmt.Errorf("MongoDB configuration is required")
		}
		if config.MongoDB.Username == "" {
			return fmt.Errorf("MongoDB username is required")
		}
		if config.MongoDB.DatabaseName == "" {
			return fmt.Errorf("MongoDB database name is required")
		}
	case databases.DatabaseTypeClickHouse:
		if config.ClickHouse == nil {
			return fmt.Errorf("ClickHouse configuration is required")
		}
		if config.ClickHouse.Username == "" {
			return fmt.Errorf("ClickHouse username is required")
		}
		if config.ClickHouse.DatabaseName == "" {
			return fmt.Errorf("ClickHouse database name is required")
		}
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	return nil
}

func (s *DatabaseService) requiresPersistentStorage(dbType databases.DatabaseType) bool {
	switch dbType {
	case databases.DatabaseTypePostgreSQL,
		databases.DatabaseTypeMySQL,
		databases.DatabaseTypeMariaDB,
		databases.DatabaseTypeMongoDB,
		databases.DatabaseTypeClickHouse:
		return true
	case databases.DatabaseTypeRedis,
		databases.DatabaseTypeKeyDB,
		databases.DatabaseTypeDragonfly:
		return false
	default:
		return false
	}
}

func (s *DatabaseService) createDefaultDisk(ctx context.Context, database *databases.Database) error {
	var mountPath string
	switch database.Type() {
	case databases.DatabaseTypePostgreSQL:
		mountPath = "/var/lib/postgresql/data"
	case databases.DatabaseTypeMySQL:
		mountPath = "/var/lib/mysql"
	case databases.DatabaseTypeMariaDB:
		mountPath = "/var/lib/mysql"
	case databases.DatabaseTypeMongoDB:
		mountPath = "/data/db"
	case databases.DatabaseTypeClickHouse:
		mountPath = "/var/lib/clickhouse"
	default:
		return fmt.Errorf("unsupported database type for disk creation: %s", database.Type())
	}

	diskName, err := disks.NewDiskName(fmt.Sprintf("%s-data", database.Name().String()))
	if err != nil {
		return fmt.Errorf("failed to create disk name: %w", err)
	}

	diskSize, err := disks.NewDiskSizeFromGB(0)
	if err != nil {
		return fmt.Errorf("failed to create disk size: %w", err)
	}

	disk, err := s.diskService.CreateDisk(
		ctx,
		diskName,
		database.ProjectID(),
		diskSize,
		mountPath,
		disks.FilesystemExt4,
		true,
	)
	if err != nil {
		return fmt.Errorf("failed to create disk: %w", err)
	}

	serviceID, err := uuid.Parse(database.ID().String())
	if err != nil {
		return fmt.Errorf("failed to parse database ID: %w", err)
	}

	if err := s.diskService.AttachDisk(ctx, disk.ID(), serviceID); err != nil {
		return fmt.Errorf("failed to attach disk: %w", err)
	}

	return nil
}
