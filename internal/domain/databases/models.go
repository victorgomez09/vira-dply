package databases

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeMariaDB    DatabaseType = "mariadb"
	DatabaseTypeRedis      DatabaseType = "redis"
	DatabaseTypeKeyDB      DatabaseType = "keydb"
	DatabaseTypeDragonfly  DatabaseType = "dragonfly"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
	DatabaseTypeClickHouse DatabaseType = "clickhouse"
)

// DatabaseStatus represents the current status of a database
type DatabaseStatus string

const (
	DatabaseStatusCreated      DatabaseStatus = "created"
	DatabaseStatusProvisioning DatabaseStatus = "provisioning"
	DatabaseStatusRunning      DatabaseStatus = "running"
	DatabaseStatusStopped      DatabaseStatus = "stopped"
	DatabaseStatusFailed       DatabaseStatus = "failed"
	DatabaseStatusDeleting     DatabaseStatus = "deleting"
)

// DatabaseConfig holds configuration specific to each database type
type DatabaseConfig struct {
	Type       DatabaseType      `json:"type"`
	PostgreSQL *PostgreSQLConfig `json:"postgresql,omitempty"`
	MySQL      *MySQLConfig      `json:"mysql,omitempty"`
	MariaDB    *MariaDBConfig    `json:"mariadb,omitempty"`
	Redis      *RedisConfig      `json:"redis,omitempty"`
	KeyDB      *KeyDBConfig      `json:"keydb,omitempty"`
	Dragonfly  *DragonflyConfig  `json:"dragonfly,omitempty"`
	MongoDB    *MongoDBConfig    `json:"mongodb,omitempty"`
	ClickHouse *ClickHouseConfig `json:"clickhouse,omitempty"`
}

// PostgreSQL configuration
type PostgreSQLConfig struct {
	Version      string            `json:"version"`
	DatabaseName string            `json:"database_name"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	Port         int               `json:"port"`
	Extensions   []string          `json:"extensions,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Resources    *ResourceConfig   `json:"resources,omitempty"`
}

// MySQL configuration
type MySQLConfig struct {
	Version      string            `json:"version"`
	DatabaseName string            `json:"database_name"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	RootPassword string            `json:"root_password"`
	Port         int               `json:"port"`
	CharacterSet string            `json:"character_set,omitempty"`
	Collation    string            `json:"collation,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Resources    *ResourceConfig   `json:"resources,omitempty"`
}

// MariaDB configuration
type MariaDBConfig struct {
	Version      string            `json:"version"`
	DatabaseName string            `json:"database_name"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	RootPassword string            `json:"root_password"`
	Port         int               `json:"port"`
	CharacterSet string            `json:"character_set,omitempty"`
	Collation    string            `json:"collation,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Resources    *ResourceConfig   `json:"resources,omitempty"`
}

// Redis configuration
type RedisConfig struct {
	Version         string            `json:"version"`
	Password        string            `json:"password,omitempty"`
	Port            int               `json:"port"`
	Database        int               `json:"database"`
	MaxMemory       string            `json:"max_memory,omitempty"`
	MaxMemoryPolicy string            `json:"max_memory_policy,omitempty"`
	Persistence     *RedisPersistence `json:"persistence,omitempty"`
	Environment     map[string]string `json:"environment,omitempty"`
	Resources       *ResourceConfig   `json:"resources,omitempty"`
}

type RedisPersistence struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"` // "rdb", "aof", or "both"
}

// KeyDB configuration
type KeyDBConfig struct {
	Version         string            `json:"version"`
	Password        string            `json:"password,omitempty"`
	Port            int               `json:"port"`
	Database        int               `json:"database"`
	MaxMemory       string            `json:"max_memory,omitempty"`
	MaxMemoryPolicy string            `json:"max_memory_policy,omitempty"`
	Persistence     *RedisPersistence `json:"persistence,omitempty"`
	Environment     map[string]string `json:"environment,omitempty"`
	Resources       *ResourceConfig   `json:"resources,omitempty"`
}

// Dragonfly configuration
type DragonflyConfig struct {
	Version     string            `json:"version"`
	Password    string            `json:"password,omitempty"`
	Port        int               `json:"port"`
	MaxMemory   string            `json:"max_memory,omitempty"`
	Persistence bool              `json:"persistence"`
	Environment map[string]string `json:"environment,omitempty"`
	Resources   *ResourceConfig   `json:"resources,omitempty"`
}

// MongoDB configuration
type MongoDBConfig struct {
	Version      string            `json:"version"`
	DatabaseName string            `json:"database_name"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	Port         int               `json:"port"`
	AuthSource   string            `json:"auth_source,omitempty"`
	ReplicaSet   string            `json:"replica_set,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Resources    *ResourceConfig   `json:"resources,omitempty"`
}

// ClickHouse configuration
type ClickHouseConfig struct {
	Version      string            `json:"version"`
	DatabaseName string            `json:"database_name"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	Port         int               `json:"port"`
	HTTPPort     int               `json:"http_port"`
	Environment  map[string]string `json:"environment,omitempty"`
	Resources    *ResourceConfig   `json:"resources,omitempty"`
}

// ResourceConfig defines resource constraints for databases
type ResourceConfig struct {
	CPULimit      string `json:"cpu_limit,omitempty"`    // e.g., "1000m" for 1 CPU
	MemoryLimit   string `json:"memory_limit,omitempty"` // e.g., "512Mi"
	CPURequest    string `json:"cpu_request,omitempty"`
	MemoryRequest string `json:"memory_request,omitempty"`
	StorageSize   string `json:"storage_size,omitempty"` // e.g., "10Gi"
}

// Database represents a database instance
type Database struct {
	id               DatabaseID
	name             DatabaseName
	description      string
	dbType           DatabaseType
	projectID        uuid.UUID
	environmentID    uuid.UUID
	config           DatabaseConfig
	status           DatabaseStatus
	connectionString string
	ports            map[string]int // map of service name to port
	containerID      string         // container ID for deployed database
	createdAt        time.Time
	updatedAt        time.Time
}

// DatabaseID value object
type DatabaseID struct {
	value string
}

func NewDatabaseID() DatabaseID {
	return DatabaseID{value: uuid.New().String()}
}

func DatabaseIDFromString(s string) (DatabaseID, error) {
	if s == "" {
		return DatabaseID{}, fmt.Errorf("database ID cannot be empty")
	}
	return DatabaseID{value: s}, nil
}

func (id DatabaseID) String() string {
	return id.value
}

// DatabaseName value object
type DatabaseName struct {
	value string
}

func NewDatabaseName(name string) (DatabaseName, error) {
	if name == "" {
		return DatabaseName{}, fmt.Errorf("database name cannot be empty")
	}
	if len(name) > 63 {
		return DatabaseName{}, fmt.Errorf("database name cannot exceed 63 characters")
	}
	return DatabaseName{value: name}, nil
}

func (n DatabaseName) String() string {
	return n.value
}

// Constructor
func NewDatabase(
	name DatabaseName,
	description string,
	dbType DatabaseType,
	projectID, environmentID uuid.UUID,
	config DatabaseConfig,
) *Database {
	now := time.Now()
	return &Database{
		id:            NewDatabaseID(),
		name:          name,
		description:   description,
		dbType:        dbType,
		projectID:     projectID,
		environmentID: environmentID,
		config:        config,
		status:        DatabaseStatusCreated,
		ports:         make(map[string]int),
		createdAt:     now,
		updatedAt:     now,
	}
}

// Getters
func (d *Database) ID() DatabaseID {
	return d.id
}

func (d *Database) Name() DatabaseName {
	return d.name
}

func (d *Database) Description() string {
	return d.description
}

func (d *Database) Type() DatabaseType {
	return d.dbType
}

func (d *Database) ProjectID() uuid.UUID {
	return d.projectID
}

func (d *Database) EnvironmentID() uuid.UUID {
	return d.environmentID
}

func (d *Database) Config() DatabaseConfig {
	return d.config
}

func (d *Database) Status() DatabaseStatus {
	return d.status
}

func (d *Database) ConnectionString() string {
	return d.connectionString
}

func (d *Database) Ports() map[string]int {
	result := make(map[string]int)
	for k, v := range d.ports {
		result[k] = v
	}
	return result
}

func (d *Database) CreatedAt() time.Time {
	return d.createdAt
}

func (d *Database) UpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Database) ContainerID() string {
	return d.containerID
}

// Setters
func (d *Database) UpdateDescription(description string) {
	d.description = description
	d.updatedAt = time.Now()
}

func (d *Database) UpdateConfig(config DatabaseConfig) {
	d.config = config
	d.updatedAt = time.Now()
}

func (d *Database) ChangeStatus(status DatabaseStatus) {
	d.status = status
	d.updatedAt = time.Now()
}

func (d *Database) SetConnectionString(connectionString string) {
	d.connectionString = connectionString
	d.updatedAt = time.Now()
}

func (d *Database) SetPorts(ports map[string]int) {
	d.ports = make(map[string]int)
	for k, v := range ports {
		d.ports[k] = v
	}
	d.updatedAt = time.Now()
}

func (d *Database) SetContainerID(containerID string) {
	d.containerID = containerID
	d.updatedAt = time.Now()
}

// Business logic methods
func (d *Database) CanStart() error {
	switch d.status {
	case DatabaseStatusRunning:
		return fmt.Errorf("database is already running")
	case DatabaseStatusProvisioning:
		return fmt.Errorf("database is currently being provisioned")
	case DatabaseStatusDeleting:
		return fmt.Errorf("database is being deleted")
	default:
		return nil
	}
}

func (d *Database) CanStop() error {
	if d.status != DatabaseStatusRunning {
		return fmt.Errorf("database is not running")
	}
	return nil
}

func (d *Database) CanDelete() error {
	switch d.status {
	case DatabaseStatusDeleting:
		return fmt.Errorf("database is already being deleted")
	case DatabaseStatusProvisioning:
		return fmt.Errorf("cannot delete database while it's being provisioned")
	default:
		return nil
	}
}

// Helper method to get the main service port
func (d *Database) GetMainPort() int {
	switch d.dbType {
	case DatabaseTypePostgreSQL:
		if d.config.PostgreSQL != nil {
			return d.config.PostgreSQL.Port
		}
		return 5432
	case DatabaseTypeMySQL:
		if d.config.MySQL != nil {
			return d.config.MySQL.Port
		}
		return 3306
	case DatabaseTypeMariaDB:
		if d.config.MariaDB != nil {
			return d.config.MariaDB.Port
		}
		return 3306
	case DatabaseTypeRedis:
		if d.config.Redis != nil {
			return d.config.Redis.Port
		}
		return 6379
	case DatabaseTypeKeyDB:
		if d.config.KeyDB != nil {
			return d.config.KeyDB.Port
		}
		return 6379
	case DatabaseTypeDragonfly:
		if d.config.Dragonfly != nil {
			return d.config.Dragonfly.Port
		}
		return 6379
	case DatabaseTypeMongoDB:
		if d.config.MongoDB != nil {
			return d.config.MongoDB.Port
		}
		return 27017
	case DatabaseTypeClickHouse:
		if d.config.ClickHouse != nil {
			return d.config.ClickHouse.Port
		}
		return 9000
	default:
		return 0
	}
}

// Helper method to generate default connection string
func (d *Database) GenerateConnectionString() string {
	switch d.dbType {
	case DatabaseTypePostgreSQL:
		if cfg := d.config.PostgreSQL; cfg != nil {
			return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
				cfg.Username, cfg.Password, "localhost", cfg.Port, cfg.DatabaseName)
		}
	case DatabaseTypeMySQL:
		if cfg := d.config.MySQL; cfg != nil {
			return fmt.Sprintf("%s:%s@tcp(localhost:%d)/%s",
				cfg.Username, cfg.Password, cfg.Port, cfg.DatabaseName)
		}
	case DatabaseTypeMariaDB:
		if cfg := d.config.MariaDB; cfg != nil {
			return fmt.Sprintf("%s:%s@tcp(localhost:%d)/%s",
				cfg.Username, cfg.Password, cfg.Port, cfg.DatabaseName)
		}
	case DatabaseTypeRedis:
		if cfg := d.config.Redis; cfg != nil {
			if cfg.Password != "" {
				return fmt.Sprintf("redis://:%s@localhost:%d/%d", cfg.Password, cfg.Port, cfg.Database)
			}
			return fmt.Sprintf("redis://localhost:%d/%d", cfg.Port, cfg.Database)
		}
	case DatabaseTypeKeyDB:
		if cfg := d.config.KeyDB; cfg != nil {
			if cfg.Password != "" {
				return fmt.Sprintf("redis://:%s@localhost:%d/%d", cfg.Password, cfg.Port, cfg.Database)
			}
			return fmt.Sprintf("redis://localhost:%d/%d", cfg.Port, cfg.Database)
		}
	case DatabaseTypeDragonfly:
		if cfg := d.config.Dragonfly; cfg != nil {
			if cfg.Password != "" {
				return fmt.Sprintf("redis://:%s@localhost:%d", cfg.Password, cfg.Port)
			}
			return fmt.Sprintf("redis://localhost:%d", cfg.Port)
		}
	case DatabaseTypeMongoDB:
		if cfg := d.config.MongoDB; cfg != nil {
			return fmt.Sprintf("mongodb://%s:%s@localhost:%d/%s",
				cfg.Username, cfg.Password, cfg.Port, cfg.DatabaseName)
		}
	case DatabaseTypeClickHouse:
		if cfg := d.config.ClickHouse; cfg != nil {
			return fmt.Sprintf("tcp://localhost:%d?database=%s&username=%s&password=%s",
				cfg.Port, cfg.DatabaseName, cfg.Username, cfg.Password)
		}
	}
	return ""
}

// Reconstruction helper for repository layer
func ReconstructDatabase(
	id DatabaseID,
	name DatabaseName,
	description string,
	dbType DatabaseType,
	projectID, environmentID uuid.UUID,
	config DatabaseConfig,
	status DatabaseStatus,
	connectionString string,
	ports map[string]int,
	containerID string,
	createdAt, updatedAt time.Time,
) *Database {
	if ports == nil {
		ports = make(map[string]int)
	}
	return &Database{
		id:               id,
		name:             name,
		description:      description,
		dbType:           dbType,
		projectID:        projectID,
		environmentID:    environmentID,
		config:           config,
		status:           status,
		connectionString: connectionString,
		ports:            ports,
		containerID:      containerID,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

// Default configuration generators
func DefaultPostgreSQLConfig() *PostgreSQLConfig {
	return &PostgreSQLConfig{
		Version:      "16",
		DatabaseName: "postgres",
		Username:     "postgres",
		Password:     generatePassword(),
		Port:         5432,
		Extensions:   []string{},
		Environment:  make(map[string]string),
		Resources: &ResourceConfig{
			MemoryLimit: "512Mi",
			CPULimit:    "500m",
			StorageSize: "10Gi",
		},
	}
}

func DefaultMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Version:      "8.0",
		DatabaseName: "mysql",
		Username:     "mysql",
		Password:     generatePassword(),
		RootPassword: generatePassword(),
		Port:         3306,
		CharacterSet: "utf8mb4",
		Collation:    "utf8mb4_unicode_ci",
		Environment:  make(map[string]string),
		Resources: &ResourceConfig{
			MemoryLimit: "512Mi",
			CPULimit:    "500m",
			StorageSize: "10Gi",
		},
	}
}

func DefaultMariaDBConfig() *MariaDBConfig {
	return &MariaDBConfig{
		Version:      "10.11",
		DatabaseName: "mariadb",
		Username:     "mariadb",
		Password:     generatePassword(),
		RootPassword: generatePassword(),
		Port:         3306,
		CharacterSet: "utf8mb4",
		Collation:    "utf8mb4_unicode_ci",
		Environment:  make(map[string]string),
		Resources: &ResourceConfig{
			MemoryLimit: "512Mi",
			CPULimit:    "500m",
			StorageSize: "10Gi",
		},
	}
}

func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Version:         "7",
		Password:        generatePassword(),
		Port:            6379,
		Database:        0,
		MaxMemory:       "256mb",
		MaxMemoryPolicy: "allkeys-lru",
		Persistence: &RedisPersistence{
			Enabled: true,
			Type:    "rdb",
		},
		Environment: make(map[string]string),
		Resources: &ResourceConfig{
			MemoryLimit: "512Mi",
			CPULimit:    "500m",
			StorageSize: "5Gi",
		},
	}
}

func DefaultKeyDBConfig() *KeyDBConfig {
	return &KeyDBConfig{
		Version:         "6.3",
		Password:        generatePassword(),
		Port:            6379,
		Database:        0,
		MaxMemory:       "256mb",
		MaxMemoryPolicy: "allkeys-lru",
		Persistence: &RedisPersistence{
			Enabled: true,
			Type:    "rdb",
		},
		Environment: make(map[string]string),
		Resources: &ResourceConfig{
			MemoryLimit: "512Mi",
			CPULimit:    "500m",
			StorageSize: "5Gi",
		},
	}
}

func DefaultDragonflyConfig() *DragonflyConfig {
	return &DragonflyConfig{
		Version:     "1.0",
		Password:    generatePassword(),
		Port:        6379,
		MaxMemory:   "256mb",
		Persistence: true,
		Environment: make(map[string]string),
		Resources: &ResourceConfig{
			MemoryLimit: "512Mi",
			CPULimit:    "500m",
			StorageSize: "5Gi",
		},
	}
}

func DefaultMongoDBConfig() *MongoDBConfig {
	return &MongoDBConfig{
		Version:      "7.0",
		DatabaseName: "mongodb",
		Username:     "mongodb",
		Password:     generatePassword(),
		Port:         27017,
		AuthSource:   "admin",
		Environment:  make(map[string]string),
		Resources: &ResourceConfig{
			MemoryLimit: "1Gi",
			CPULimit:    "1000m",
			StorageSize: "20Gi",
		},
	}
}

func DefaultClickHouseConfig() *ClickHouseConfig {
	return &ClickHouseConfig{
		Version:      "23.8",
		DatabaseName: "default",
		Username:     "default",
		Password:     generatePassword(),
		Port:         9000,
		HTTPPort:     8123,
		Environment:  make(map[string]string),
		Resources: &ResourceConfig{
			MemoryLimit: "2Gi",
			CPULimit:    "1000m",
			StorageSize: "50Gi",
		},
	}
}

// Helper function to generate secure passwords
func generatePassword() string {
	// Simple password generation - in production, use crypto/rand
	return uuid.New().String()[:16]
}

// Helper function to marshal config to JSON for storage
func (d *Database) ConfigJSON() ([]byte, error) {
	return json.Marshal(d.config)
}

// Helper function to unmarshal config from JSON
func (d *Database) SetConfigFromJSON(data []byte) error {
	return json.Unmarshal(data, &d.config)
}
