package studio

import (
	"context"
	"time"
)

type ColumnType string

const (
	ColumnTypeString    ColumnType = "string"
	ColumnTypeInteger   ColumnType = "integer"
	ColumnTypeFloat     ColumnType = "float"
	ColumnTypeBoolean   ColumnType = "boolean"
	ColumnTypeDate      ColumnType = "date"
	ColumnTypeDateTime  ColumnType = "datetime"
	ColumnTypeTimestamp ColumnType = "timestamp"
	ColumnTypeJSON      ColumnType = "json"
	ColumnTypeBinary    ColumnType = "binary"
	ColumnTypeText      ColumnType = "text"
	ColumnTypeUUID      ColumnType = "uuid"
	ColumnTypeArray     ColumnType = "array"
	ColumnTypeUnknown   ColumnType = "unknown"
)

type Column struct {
	Name         string     `json:"name"`
	Type         ColumnType `json:"type"`
	NativeType   string     `json:"native_type"`
	Nullable     bool       `json:"nullable"`
	DefaultValue *string    `json:"default_value,omitempty"`
	IsPrimaryKey bool       `json:"is_primary_key"`
	IsUnique     bool       `json:"is_unique"`
	IsAutoInc    bool       `json:"is_auto_increment"`
	MaxLength    *int       `json:"max_length,omitempty"`
}

type ForeignKey struct {
	Name             string `json:"name"`
	Column           string `json:"column"`
	ReferencedTable  string `json:"referenced_table"`
	ReferencedColumn string `json:"referenced_column"`
	OnDelete         string `json:"on_delete"`
	OnUpdate         string `json:"on_update"`
}

type Index struct {
	Name     string   `json:"name"`
	Columns  []string `json:"columns"`
	IsUnique bool     `json:"is_unique"`
	Type     string   `json:"type"`
}

type TableSchema struct {
	Name        string       `json:"name"`
	Schema      string       `json:"schema,omitempty"`
	Type        string       `json:"type"`
	Columns     []Column     `json:"columns"`
	PrimaryKeys []string     `json:"primary_keys"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
	Indexes     []Index      `json:"indexes"`
	RowCount    *int64       `json:"row_count,omitempty"`
	Comment     string       `json:"comment,omitempty"`
}

type QueryResult struct {
	Columns       []string         `json:"columns"`
	Rows          []map[string]any `json:"rows"`
	RowsAffected  int64            `json:"rows_affected"`
	ExecutionTime time.Duration    `json:"execution_time"`
	TotalCount    *int64           `json:"total_count,omitempty"`
	Error         string           `json:"error,omitempty"`
}

type QueryOptions struct {
	Limit  int
	Offset int
}

type Filter struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type Sort struct {
	Column    string `json:"column"`
	Direction string `json:"direction"`
}

type TableDataOptions struct {
	Filters []Filter
	Sorts   []Sort
	Limit   int
	Offset  int
}

type DatabaseClient interface {
	Connect(ctx context.Context, connectionString string) error
	Close() error

	ListDatabases(ctx context.Context) ([]string, error)
	ListSchemas(ctx context.Context) ([]string, error)
	ListTables(ctx context.Context, schema string) ([]string, error)
	GetTableSchema(ctx context.Context, tableName, schema string) (*TableSchema, error)
	GetTableData(ctx context.Context, tableName, schema string, opts TableDataOptions) (*QueryResult, error)

	ExecuteQuery(ctx context.Context, query string, opts QueryOptions) (*QueryResult, error)

	InsertRow(ctx context.Context, tableName, schema string, data map[string]any) error
	UpdateRow(ctx context.Context, tableName, schema string, primaryKey map[string]any, data map[string]any) error
	DeleteRow(ctx context.Context, tableName, schema string, primaryKey map[string]any) error

	GetDatabaseVersion(ctx context.Context) (string, error)
	GetDatabaseSize(ctx context.Context) (int64, error)
}
