package queue_db

import (
	"fmt"
)

// QueueFactory implements the Factory interface
type QueueFactory struct{}

// NewQueueFactory creates a new queue database factory
func NewQueueFactory() Factory {
	return &QueueFactory{}
}

// Create creates a queue database instance based on configuration
func (f *QueueFactory) Create(dbType DatabaseType, connectionString string) (QueueDatabase, error) {
	switch dbType {
	case Dragonfly:
		return NewDragonflyDatabase(connectionString)
	case Redis:
		return NewDragonflyDatabase(connectionString) // Same implementation works for Redis
	default:
		return nil, fmt.Errorf("unsupported queue database type: %s", dbType)
	}
}
