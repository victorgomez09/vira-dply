package analytics_db

import (
	"fmt"
)

// AnalyticsFactory implements the Factory interface
type AnalyticsFactory struct{}

// NewAnalyticsFactory creates a new analytics database factory
func NewAnalyticsFactory() Factory {
	return &AnalyticsFactory{}
}

// Create creates an analytics database instance based on configuration
func (f *AnalyticsFactory) Create(dbType DatabaseType, connectionString string) (AnalyticsDatabase, error) {
	switch dbType {
	case SQLite:
		return NewSQLiteDatabase(connectionString)
	case DuckDB:
		return NewDuckDBDatabase(connectionString)
	case ClickHouse:
		return nil, fmt.Errorf("ClickHouse implementation not yet available")
	default:
		return nil, fmt.Errorf("unsupported analytics database type: %s", dbType)
	}
}
