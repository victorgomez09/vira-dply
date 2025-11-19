package main_db

import (
	"fmt"
)

// DatabaseFactory implements the Factory interface
type DatabaseFactory struct{}

// NewDatabaseFactory creates a new database factory
func NewDatabaseFactory() Factory {
	return &DatabaseFactory{}
}

// Create creates a main database instance based on configuration
func (f *DatabaseFactory) Create(dbType DatabaseType, connectionString string) (MainDatabase, error) {
	switch dbType {
	case SQLite:
		return NewSQLiteDatabase(connectionString)
	case PostgreSQL:
		return NewPostgreSQLDatabase(connectionString)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
