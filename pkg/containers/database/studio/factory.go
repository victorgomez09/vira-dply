package studio

import (
	"context"
	"fmt"
	"strings"

	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
)

type ClientFactory struct{}

func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

func (f *ClientFactory) CreateClient(ctx context.Context, db *databases.Database) (DatabaseClient, error) {
	switch db.Type() {
	case databases.DatabaseTypePostgreSQL:
		client := NewPostgreSQLClient()
		connStr := db.ConnectionString()
		if connStr == "" {
			connStr = db.GenerateConnectionString()
		} else {
			connStr = ensureSSLMode(connStr)
		}
		if err := client.Connect(ctx, connStr); err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		return client, nil

	case databases.DatabaseTypeMySQL, databases.DatabaseTypeMariaDB:
		client := NewMySQLClient()
		connStr := db.ConnectionString()
		if connStr == "" {
			connStr = db.GenerateConnectionString()
		}
		if err := client.Connect(ctx, connStr); err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL/MariaDB: %w", err)
		}
		return client, nil

	case databases.DatabaseTypeMongoDB:
		return nil, fmt.Errorf("MongoDB support is coming soon")

	case databases.DatabaseTypeRedis, databases.DatabaseTypeKeyDB, databases.DatabaseTypeDragonfly:
		return nil, fmt.Errorf("Redis-based databases do not support studio features (use CLI instead)")

	case databases.DatabaseTypeClickHouse:
		return nil, fmt.Errorf("ClickHouse support is coming soon")

	default:
		return nil, fmt.Errorf("unsupported database type: %s", db.Type())
	}
}

func (f *ClientFactory) SupportsStudio(dbType databases.DatabaseType) bool {
	switch dbType {
	case databases.DatabaseTypePostgreSQL, databases.DatabaseTypeMySQL, databases.DatabaseTypeMariaDB:
		return true
	default:
		return false
	}
}

func ensureSSLMode(connStr string) string {
	if !strings.Contains(connStr, "sslmode=") {
		if strings.Contains(connStr, "?") {
			return connStr + "&sslmode=disable"
		}
		return connStr + "?sslmode=disable"
	}
	return connStr
}
