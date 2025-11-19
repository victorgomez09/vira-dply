package database

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
)

// DefaultContainerConfigBuilder builds container configurations for different database types
type DefaultContainerConfigBuilder struct {
	imageResolver DatabaseImageResolver
}

func NewDefaultContainerConfigBuilder(imageResolver DatabaseImageResolver) ContainerConfigBuilder {
	return &DefaultContainerConfigBuilder{
		imageResolver: imageResolver,
	}
}

// BuildConfig builds container configuration for any database type
func (b *DefaultContainerConfigBuilder) BuildConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	dbType := database.Type()

	switch dbType {
	case databases.DatabaseTypePostgreSQL:
		return b.BuildPostgreSQLConfig(database)
	case databases.DatabaseTypeMySQL:
		return b.BuildMySQLConfig(database)
	case databases.DatabaseTypeMariaDB:
		return b.BuildMariaDBConfig(database)
	case databases.DatabaseTypeRedis:
		return b.BuildRedisConfig(database)
	case databases.DatabaseTypeKeyDB:
		return b.BuildKeyDBConfig(database)
	case databases.DatabaseTypeDragonfly:
		return b.BuildDragonflyConfig(database)
	case databases.DatabaseTypeMongoDB:
		return b.BuildMongoDBConfig(database)
	case databases.DatabaseTypeClickHouse:
		return b.BuildClickHouseConfig(database)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// BuildPostgreSQLConfig builds configuration for PostgreSQL containers
func (b *DefaultContainerConfigBuilder) BuildPostgreSQLConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	// Extract PostgreSQL config from database
	var pgConfig *databases.PostgreSQLConfig
	configBytes, err := json.Marshal(database.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig databases.DatabaseConfig
	if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	pgConfig = dbConfig.PostgreSQL
	if pgConfig == nil {
		return nil, fmt.Errorf("postgresql config is nil")
	}

	image := b.imageResolver.ResolveImage(databases.DatabaseTypePostgreSQL, pgConfig.Version)
	containerName := fmt.Sprintf("mikrocloud-%s-%s-%s",
		database.ProjectID(),
		database.EnvironmentID(),
		database.Name().String())

	environment := map[string]string{
		"POSTGRES_DB":       pgConfig.DatabaseName,
		"POSTGRES_USER":     pgConfig.Username,
		"POSTGRES_PASSWORD": pgConfig.Password,
		"PGDATA":            "/var/lib/postgresql/data/pgdata",
	}

	// Add custom environment variables
	for k, v := range pgConfig.Environment {
		environment[k] = v
	}

	volumes := map[string]string{
		fmt.Sprintf("mikrocloud-postgres-%s", database.ID().String()): "/var/lib/postgresql/data",
	}

	return &DatabaseContainerConfig{
		Database:      database,
		Image:         image,
		ContainerName: containerName,
		Port:          strconv.Itoa(pgConfig.Port),
		Environment:   environment,
		Volumes:       volumes,
		Command:       []string{},
		HealthCheck: &HealthCheckConfig{
			Test:     []string{"CMD-SHELL", "pg_isready -U " + pgConfig.Username + " -d " + pgConfig.DatabaseName},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}, nil
}

// BuildMySQLConfig builds configuration for MySQL containers
func (b *DefaultContainerConfigBuilder) BuildMySQLConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	// Extract MySQL config from database
	var mysqlConfig *databases.MySQLConfig
	configBytes, err := json.Marshal(database.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig databases.DatabaseConfig
	if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	mysqlConfig = dbConfig.MySQL
	if mysqlConfig == nil {
		return nil, fmt.Errorf("mysql config is nil")
	}

	image := b.imageResolver.ResolveImage(databases.DatabaseTypeMySQL, mysqlConfig.Version)
	containerName := fmt.Sprintf("mikrocloud-%s-%s-%s",
		database.ProjectID(),
		database.EnvironmentID(),
		database.Name().String())

	environment := map[string]string{
		"MYSQL_DATABASE":      mysqlConfig.DatabaseName,
		"MYSQL_USER":          mysqlConfig.Username,
		"MYSQL_PASSWORD":      mysqlConfig.Password,
		"MYSQL_ROOT_PASSWORD": mysqlConfig.RootPassword,
	}

	// Add custom environment variables
	for k, v := range mysqlConfig.Environment {
		environment[k] = v
	}

	volumes := map[string]string{
		fmt.Sprintf("mikrocloud-mysql-%s", database.ID().String()): "/var/lib/mysql",
	}

	return &DatabaseContainerConfig{
		Database:      database,
		Image:         image,
		ContainerName: containerName,
		Port:          strconv.Itoa(mysqlConfig.Port),
		Environment:   environment,
		Volumes:       volumes,
		Command:       []string{},
		HealthCheck: &HealthCheckConfig{
			Test:     []string{"CMD", "mysqladmin", "ping", "-h", "localhost"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}, nil
}

// BuildMariaDBConfig builds configuration for MariaDB containers
func (b *DefaultContainerConfigBuilder) BuildMariaDBConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	// Extract MariaDB config from database
	var mariaConfig *databases.MariaDBConfig
	configBytes, err := json.Marshal(database.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig databases.DatabaseConfig
	if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	mariaConfig = dbConfig.MariaDB
	if mariaConfig == nil {
		return nil, fmt.Errorf("mariadb config is nil")
	}

	image := b.imageResolver.ResolveImage(databases.DatabaseTypeMariaDB, mariaConfig.Version)
	containerName := fmt.Sprintf("mikrocloud-%s-%s-%s",
		database.ProjectID(),
		database.EnvironmentID(),
		database.Name().String())

	environment := map[string]string{
		"MARIADB_DATABASE":      mariaConfig.DatabaseName,
		"MARIADB_USER":          mariaConfig.Username,
		"MARIADB_PASSWORD":      mariaConfig.Password,
		"MARIADB_ROOT_PASSWORD": mariaConfig.RootPassword,
	}

	// Add custom environment variables
	for k, v := range mariaConfig.Environment {
		environment[k] = v
	}

	volumes := map[string]string{
		fmt.Sprintf("mikrocloud-mariadb-%s", database.ID().String()): "/var/lib/mysql",
	}

	return &DatabaseContainerConfig{
		Database:      database,
		Image:         image,
		ContainerName: containerName,
		Port:          strconv.Itoa(mariaConfig.Port),
		Environment:   environment,
		Volumes:       volumes,
		Command:       []string{},
		HealthCheck: &HealthCheckConfig{
			Test:     []string{"CMD", "healthcheck.sh", "--connect", "--innodb_initialized"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}, nil
}

// BuildRedisConfig builds configuration for Redis containers
func (b *DefaultContainerConfigBuilder) BuildRedisConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	// Extract Redis config from database
	var redisConfig *databases.RedisConfig
	configBytes, err := json.Marshal(database.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig databases.DatabaseConfig
	if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	redisConfig = dbConfig.Redis
	if redisConfig == nil {
		return nil, fmt.Errorf("redis config is nil")
	}

	image := b.imageResolver.ResolveImage(databases.DatabaseTypeRedis, redisConfig.Version)
	containerName := fmt.Sprintf("mikrocloud-%s-%s-%s",
		database.ProjectID(),
		database.EnvironmentID(),
		database.Name().String())

	environment := map[string]string{}

	// Add custom environment variables
	for k, v := range redisConfig.Environment {
		environment[k] = v
	}

	volumes := map[string]string{
		fmt.Sprintf("mikrocloud-redis-%s", database.ID().String()): "/data",
	}

	command := []string{"redis-server"}
	if redisConfig.Password != "" {
		command = append(command, "--requirepass", redisConfig.Password)
	}

	return &DatabaseContainerConfig{
		Database:      database,
		Image:         image,
		ContainerName: containerName,
		Port:          strconv.Itoa(redisConfig.Port),
		Environment:   environment,
		Volumes:       volumes,
		Command:       command,
		HealthCheck: &HealthCheckConfig{
			Test:     []string{"CMD", "redis-cli", "ping"},
			Interval: "30s",
			Timeout:  "3s",
			Retries:  3,
		},
	}, nil
}

// BuildKeyDBConfig builds configuration for KeyDB containers
func (b *DefaultContainerConfigBuilder) BuildKeyDBConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	// Extract KeyDB config from database
	var keydbConfig *databases.KeyDBConfig
	configBytes, err := json.Marshal(database.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig databases.DatabaseConfig
	if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	keydbConfig = dbConfig.KeyDB
	if keydbConfig == nil {
		return nil, fmt.Errorf("keydb config is nil")
	}

	image := b.imageResolver.ResolveImage(databases.DatabaseTypeKeyDB, keydbConfig.Version)
	containerName := fmt.Sprintf("mikrocloud-%s-%s-%s",
		database.ProjectID(),
		database.EnvironmentID(),
		database.Name().String())

	environment := map[string]string{}

	// Add custom environment variables
	for k, v := range keydbConfig.Environment {
		environment[k] = v
	}

	volumes := map[string]string{
		fmt.Sprintf("mikrocloud-keydb-%s", database.ID().String()): "/data",
	}

	command := []string{"keydb-server"}
	if keydbConfig.Password != "" {
		command = append(command, "--requirepass", keydbConfig.Password)
	}

	return &DatabaseContainerConfig{
		Database:      database,
		Image:         image,
		ContainerName: containerName,
		Port:          strconv.Itoa(keydbConfig.Port),
		Environment:   environment,
		Volumes:       volumes,
		Command:       command,
		HealthCheck: &HealthCheckConfig{
			Test:     []string{"CMD", "keydb-cli", "ping"},
			Interval: "30s",
			Timeout:  "3s",
			Retries:  3,
		},
	}, nil
}

// BuildDragonflyConfig builds configuration for Dragonfly containers
func (b *DefaultContainerConfigBuilder) BuildDragonflyConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	// Extract Dragonfly config from database
	var dragonflyConfig *databases.DragonflyConfig
	configBytes, err := json.Marshal(database.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig databases.DatabaseConfig
	if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	dragonflyConfig = dbConfig.Dragonfly
	if dragonflyConfig == nil {
		return nil, fmt.Errorf("dragonfly config is nil")
	}

	image := b.imageResolver.ResolveImage(databases.DatabaseTypeDragonfly, dragonflyConfig.Version)
	containerName := fmt.Sprintf("mikrocloud-%s-%s-%s",
		database.ProjectID(),
		database.EnvironmentID(),
		database.Name().String())

	environment := map[string]string{}

	// Add custom environment variables
	for k, v := range dragonflyConfig.Environment {
		environment[k] = v
	}

	volumes := map[string]string{
		fmt.Sprintf("mikrocloud-dragonfly-%s", database.ID().String()): "/data",
	}

	command := []string{"dragonfly", "--logtostderr"}
	if dragonflyConfig.Password != "" {
		command = append(command, "--requirepass", dragonflyConfig.Password)
	}

	return &DatabaseContainerConfig{
		Database:      database,
		Image:         image,
		ContainerName: containerName,
		Port:          strconv.Itoa(dragonflyConfig.Port),
		Environment:   environment,
		Volumes:       volumes,
		Command:       command,
		HealthCheck: &HealthCheckConfig{
			Test:     []string{"CMD", "redis-cli", "ping"},
			Interval: "30s",
			Timeout:  "3s",
			Retries:  3,
		},
	}, nil
}

// BuildMongoDBConfig builds configuration for MongoDB containers
func (b *DefaultContainerConfigBuilder) BuildMongoDBConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	// Extract MongoDB config from database
	var mongoConfig *databases.MongoDBConfig
	configBytes, err := json.Marshal(database.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig databases.DatabaseConfig
	if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	mongoConfig = dbConfig.MongoDB
	if mongoConfig == nil {
		return nil, fmt.Errorf("mongodb config is nil")
	}

	image := b.imageResolver.ResolveImage(databases.DatabaseTypeMongoDB, mongoConfig.Version)
	containerName := fmt.Sprintf("mikrocloud-%s-%s-%s",
		database.ProjectID(),
		database.EnvironmentID(),
		database.Name().String())

	environment := map[string]string{
		"MONGO_INITDB_DATABASE":      mongoConfig.DatabaseName,
		"MONGO_INITDB_ROOT_USERNAME": mongoConfig.Username,
		"MONGO_INITDB_ROOT_PASSWORD": mongoConfig.Password,
	}

	// Add custom environment variables
	for k, v := range mongoConfig.Environment {
		environment[k] = v
	}

	volumes := map[string]string{
		fmt.Sprintf("mikrocloud-mongodb-%s", database.ID().String()): "/data/db",
	}

	return &DatabaseContainerConfig{
		Database:      database,
		Image:         image,
		ContainerName: containerName,
		Port:          strconv.Itoa(mongoConfig.Port),
		Environment:   environment,
		Volumes:       volumes,
		Command:       []string{},
		HealthCheck: &HealthCheckConfig{
			Test:     []string{"CMD", "mongo", "--eval", "db.adminCommand('ping')"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}, nil
}

// BuildClickHouseConfig builds configuration for ClickHouse containers
func (b *DefaultContainerConfigBuilder) BuildClickHouseConfig(database *databases.Database) (*DatabaseContainerConfig, error) {
	// Extract ClickHouse config from database
	var clickhouseConfig *databases.ClickHouseConfig
	configBytes, err := json.Marshal(database.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig databases.DatabaseConfig
	if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	clickhouseConfig = dbConfig.ClickHouse
	if clickhouseConfig == nil {
		return nil, fmt.Errorf("clickhouse config is nil")
	}

	image := b.imageResolver.ResolveImage(databases.DatabaseTypeClickHouse, clickhouseConfig.Version)
	containerName := fmt.Sprintf("mikrocloud-%s-%s-%s",
		database.ProjectID(),
		database.EnvironmentID(),
		database.Name().String())

	environment := map[string]string{
		"CLICKHOUSE_DB":       clickhouseConfig.DatabaseName,
		"CLICKHOUSE_USER":     clickhouseConfig.Username,
		"CLICKHOUSE_PASSWORD": clickhouseConfig.Password,
	}

	// Add custom environment variables
	for k, v := range clickhouseConfig.Environment {
		environment[k] = v
	}

	volumes := map[string]string{
		fmt.Sprintf("mikrocloud-clickhouse-%s", database.ID().String()): "/var/lib/clickhouse",
	}

	return &DatabaseContainerConfig{
		Database:      database,
		Image:         image,
		ContainerName: containerName,
		Port:          strconv.Itoa(clickhouseConfig.Port),
		Environment:   environment,
		Volumes:       volumes,
		Command:       []string{},
		HealthCheck: &HealthCheckConfig{
			Test:     []string{"CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8123/ping"},
			Interval: "30s",
			Timeout:  "5s",
			Retries:  3,
		},
	}, nil
}
