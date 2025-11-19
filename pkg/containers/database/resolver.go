package database

import (
	"fmt"

	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
)

// DefaultImageResolver provides default Docker images for database types
type DefaultImageResolver struct{}

func NewDefaultImageResolver() DatabaseImageResolver {
	return &DefaultImageResolver{}
}

// ResolveImage returns the appropriate Docker image for a database type and version
func (r *DefaultImageResolver) ResolveImage(dbType databases.DatabaseType, version string) string {
	if version == "" {
		version = r.GetDefaultVersion(dbType)
	}

	switch dbType {
	case databases.DatabaseTypePostgreSQL:
		return fmt.Sprintf("postgres:%s", version)
	case databases.DatabaseTypeMySQL:
		return fmt.Sprintf("mysql:%s", version)
	case databases.DatabaseTypeMariaDB:
		return fmt.Sprintf("mariadb:%s", version)
	case databases.DatabaseTypeRedis:
		return fmt.Sprintf("redis:%s", version)
	case databases.DatabaseTypeKeyDB:
		return fmt.Sprintf("eqalpha/keydb:%s", version)
	case databases.DatabaseTypeDragonfly:
		return fmt.Sprintf("dragonflydb/dragonfly:%s", version)
	case databases.DatabaseTypeMongoDB:
		return fmt.Sprintf("mongo:%s", version)
	case databases.DatabaseTypeClickHouse:
		return fmt.Sprintf("clickhouse/clickhouse-server:%s", version)
	default:
		return fmt.Sprintf("postgres:%s", r.GetDefaultVersion(databases.DatabaseTypePostgreSQL))
	}
}

// GetDefaultVersion returns the default version for each database type
func (r *DefaultImageResolver) GetDefaultVersion(dbType databases.DatabaseType) string {
	switch dbType {
	case databases.DatabaseTypePostgreSQL:
		return "16-alpine"
	case databases.DatabaseTypeMySQL:
		return "8.4"
	case databases.DatabaseTypeMariaDB:
		return "11.6"
	case databases.DatabaseTypeRedis:
		return "7.4-alpine"
	case databases.DatabaseTypeKeyDB:
		return "v6.3.4"
	case databases.DatabaseTypeDragonfly:
		return "v1.23.1"
	case databases.DatabaseTypeMongoDB:
		return "8.0"
	case databases.DatabaseTypeClickHouse:
		return "24.12.1.1823"
	default:
		return "16-alpine"
	}
}

// GetSupportedVersions returns a list of supported versions for each database type
func (r *DefaultImageResolver) GetSupportedVersions(dbType databases.DatabaseType) []string {
	switch dbType {
	case databases.DatabaseTypePostgreSQL:
		return []string{"16-alpine", "15-alpine", "14-alpine", "13-alpine"}
	case databases.DatabaseTypeMySQL:
		return []string{"8.4", "8.0", "5.7"}
	case databases.DatabaseTypeMariaDB:
		return []string{"11.6", "10.11", "10.6"}
	case databases.DatabaseTypeRedis:
		return []string{"7.4-alpine", "7.2-alpine", "6.2-alpine"}
	case databases.DatabaseTypeKeyDB:
		return []string{"v6.3.4", "v6.3.3", "v6.3.2"}
	case databases.DatabaseTypeDragonfly:
		return []string{"v1.23.1", "v1.22.2", "v1.21.1"}
	case databases.DatabaseTypeMongoDB:
		return []string{"8.0", "7.0", "6.0"}
	case databases.DatabaseTypeClickHouse:
		return []string{"24.12.1.1823", "24.11.1.2557", "24.10.2.80"}
	default:
		return []string{"16-alpine"}
	}
}
