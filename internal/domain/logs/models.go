package logs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	id        LogEntryID
	projectID uuid.UUID
	serviceID *uuid.UUID
	level     LogLevel
	message   string
	timestamp time.Time
	source    LogSource
	metadata  map[string]interface{}
	createdAt time.Time
}

type LogEntryID struct {
	value string
}

func NewLogEntryID() LogEntryID {
	return LogEntryID{value: uuid.New().String()}
}

func LogEntryIDFromString(s string) (LogEntryID, error) {
	if s == "" {
		return LogEntryID{}, fmt.Errorf("log entry ID cannot be empty")
	}
	return LogEntryID{value: s}, nil
}

func (id LogEntryID) String() string {
	return id.value
}

type LogLevel string

const (
	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

type LogSource string

const (
	LogSourceApplication LogSource = "application"
	LogSourceSystem      LogSource = "system"
	LogSourceContainer   LogSource = "container"
	LogSourceProxy       LogSource = "proxy"
	LogSourceBuild       LogSource = "build"
	LogSourceDeploy      LogSource = "deploy"
)

type LogQuery struct {
	ProjectID uuid.UUID
	ServiceID *uuid.UUID
	Level     *LogLevel
	Source    *LogSource
	StartTime *time.Time
	EndTime   *time.Time
	Search    string
	Limit     int
	Offset    int
}

func NewLogEntry(
	projectID uuid.UUID,
	serviceID *uuid.UUID,
	level LogLevel,
	message string,
	source LogSource,
	metadata map[string]interface{},
) *LogEntry {
	now := time.Now()
	return &LogEntry{
		id:        NewLogEntryID(),
		projectID: projectID,
		serviceID: serviceID,
		level:     level,
		message:   message,
		timestamp: now,
		source:    source,
		metadata:  metadata,
		createdAt: now,
	}
}

func (l *LogEntry) ID() LogEntryID {
	return l.id
}

func (l *LogEntry) ProjectID() uuid.UUID {
	return l.projectID
}

func (l *LogEntry) ServiceID() *uuid.UUID {
	return l.serviceID
}

func (l *LogEntry) Level() LogLevel {
	return l.level
}

func (l *LogEntry) Message() string {
	return l.message
}

func (l *LogEntry) Timestamp() time.Time {
	return l.timestamp
}

func (l *LogEntry) Source() LogSource {
	return l.source
}

func (l *LogEntry) Metadata() map[string]interface{} {
	return l.metadata
}

func (l *LogEntry) CreatedAt() time.Time {
	return l.createdAt
}

func ReconstructLogEntry(
	id LogEntryID,
	projectID uuid.UUID,
	serviceID *uuid.UUID,
	level LogLevel,
	message string,
	timestamp time.Time,
	source LogSource,
	metadata map[string]interface{},
	createdAt time.Time,
) *LogEntry {
	return &LogEntry{
		id:        id,
		projectID: projectID,
		serviceID: serviceID,
		level:     level,
		message:   message,
		timestamp: timestamp,
		source:    source,
		metadata:  metadata,
		createdAt: createdAt,
	}
}
