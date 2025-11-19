package studio

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type PostgreSQLClient struct {
	db *sql.DB
}

func NewPostgreSQLClient() *PostgreSQLClient {
	return &PostgreSQLClient{}
}

func (c *PostgreSQLClient) Connect(ctx context.Context, connectionString string) error {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.db = db
	return nil
}

func (c *PostgreSQLClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *PostgreSQLClient) ListDatabases(ctx context.Context) ([]string, error) {
	query := `SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname`
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		databases = append(databases, name)
	}

	return databases, rows.Err()
}

func (c *PostgreSQLClient) ListSchemas(ctx context.Context) ([]string, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY schema_name
	`
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		schemas = append(schemas, name)
	}

	return schemas, rows.Err()
}

func (c *PostgreSQLClient) ListTables(ctx context.Context, schema string) ([]string, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = $1 
		ORDER BY tablename
	`
	rows, err := c.db.QueryContext(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	return tables, rows.Err()
}

func (c *PostgreSQLClient) GetTableSchema(ctx context.Context, tableName, schema string) (*TableSchema, error) {
	if schema == "" {
		schema = "public"
	}

	tableSchema := &TableSchema{
		Name:   tableName,
		Schema: schema,
		Type:   "table",
	}

	columnsQuery := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			COALESCE(tc.constraint_type, '') as constraint_type,
			CASE WHEN c.column_default LIKE 'nextval%' THEN true ELSE false END as is_auto_inc
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu 
			ON c.table_schema = kcu.table_schema 
			AND c.table_name = kcu.table_name 
			AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc 
			ON kcu.constraint_name = tc.constraint_name
			AND kcu.table_schema = tc.table_schema
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`

	rows, err := c.db.QueryContext(ctx, columnsQuery, schema, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col Column
		var defaultVal sql.NullString
		var maxLength sql.NullInt64
		var constraintType string
		var nativeType string
		var nullable string

		if err := rows.Scan(&col.Name, &nativeType, &nullable, &defaultVal, &maxLength, &constraintType, &col.IsAutoInc); err != nil {
			return nil, err
		}

		col.NativeType = nativeType
		col.Type = mapPostgresType(nativeType)
		col.Nullable = nullable == "YES"

		if defaultVal.Valid {
			col.DefaultValue = &defaultVal.String
		}

		if maxLength.Valid {
			maxLen := int(maxLength.Int64)
			col.MaxLength = &maxLen
		}

		col.IsPrimaryKey = constraintType == "PRIMARY KEY"
		col.IsUnique = constraintType == "UNIQUE" || col.IsPrimaryKey

		if col.IsPrimaryKey {
			tableSchema.PrimaryKeys = append(tableSchema.PrimaryKeys, col.Name)
		}

		tableSchema.Columns = append(tableSchema.Columns, col)
	}

	fkQuery := `
		SELECT 
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name as foreign_table_name,
			ccu.column_name as foreign_column_name,
			rc.update_rule,
			rc.delete_rule
		FROM information_schema.table_constraints AS tc 
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		JOIN information_schema.referential_constraints AS rc
			ON tc.constraint_name = rc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY' 
			AND tc.table_schema = $1
			AND tc.table_name = $2
	`

	fkRows, err := c.db.QueryContext(ctx, fkQuery, schema, tableName)
	if err != nil {
		return nil, err
	}
	defer fkRows.Close()

	for fkRows.Next() {
		var fk ForeignKey
		if err := fkRows.Scan(&fk.Name, &fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn, &fk.OnUpdate, &fk.OnDelete); err != nil {
			return nil, err
		}
		tableSchema.ForeignKeys = append(tableSchema.ForeignKeys, fk)
	}

	indexQuery := `
		SELECT 
			i.relname as index_name,
			a.attname as column_name,
			ix.indisunique as is_unique,
			am.amname as index_type
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_am am ON i.relam = am.oid
		JOIN pg_namespace n ON t.relnamespace = n.oid
		WHERE n.nspname = $1 AND t.relname = $2
		ORDER BY i.relname, a.attnum
	`

	indexRows, err := c.db.QueryContext(ctx, indexQuery, schema, tableName)
	if err != nil {
		return nil, err
	}
	defer indexRows.Close()

	indexMap := make(map[string]*Index)
	for indexRows.Next() {
		var indexName, columnName, indexType string
		var isUnique bool

		if err := indexRows.Scan(&indexName, &columnName, &isUnique, &indexType); err != nil {
			return nil, err
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &Index{
				Name:     indexName,
				Columns:  []string{columnName},
				IsUnique: isUnique,
				Type:     indexType,
			}
		}
	}

	for _, idx := range indexMap {
		tableSchema.Indexes = append(tableSchema.Indexes, *idx)
	}

	var rowCount int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.%s`, schema, tableName)
	if err := c.db.QueryRowContext(ctx, countQuery).Scan(&rowCount); err == nil {
		tableSchema.RowCount = &rowCount
	}

	return tableSchema, nil
}

func (c *PostgreSQLClient) GetTableData(ctx context.Context, tableName, schema string, opts TableDataOptions) (*QueryResult, error) {
	if schema == "" {
		schema = "public"
	}

	start := time.Now()

	query := fmt.Sprintf(`SELECT * FROM %s.%s`, schema, tableName)

	whereClauses := []string{}
	args := []any{}
	argIndex := 1

	for _, filter := range opts.Filters {
		clause, arg := buildWhereClause(filter, argIndex)
		whereClauses = append(whereClauses, clause)
		args = append(args, arg)
		argIndex++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if len(opts.Sorts) > 0 {
		orderParts := []string{}
		for _, sort := range opts.Sorts {
			direction := "ASC"
			if strings.ToUpper(sort.Direction) == "DESC" {
				direction = "DESC"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", sort.Column, direction))
		}
		query += " ORDER BY " + strings.Join(orderParts, ", ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.%s`, schema, tableName)
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var totalCount int64
	if err := c.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, err
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}
	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return &QueryResult{Error: err.Error()}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var resultRows []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}
		resultRows = append(resultRows, row)
	}

	return &QueryResult{
		Columns:       columns,
		Rows:          resultRows,
		RowsAffected:  int64(len(resultRows)),
		ExecutionTime: time.Since(start),
		TotalCount:    &totalCount,
	}, rows.Err()
}

func (c *PostgreSQLClient) ExecuteQuery(ctx context.Context, query string, opts QueryOptions) (*QueryResult, error) {
	start := time.Now()

	trimmedQuery := strings.TrimSpace(strings.ToUpper(query))
	isSelect := strings.HasPrefix(trimmedQuery, "SELECT")

	if isSelect {
		hasLimit := strings.Contains(trimmedQuery, "LIMIT")
		hasOffset := strings.Contains(trimmedQuery, "OFFSET")

		if opts.Limit > 0 && !hasLimit {
			query = fmt.Sprintf("%s LIMIT %d", query, opts.Limit)
		}
		if opts.Offset > 0 && !hasOffset {
			query = fmt.Sprintf("%s OFFSET %d", query, opts.Offset)
		}

		rows, err := c.db.QueryContext(ctx, query)
		if err != nil {
			return &QueryResult{Error: err.Error()}, err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		var resultRows []map[string]any
		for rows.Next() {
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}

			row := make(map[string]any)
			for i, col := range columns {
				row[col] = values[i]
			}
			resultRows = append(resultRows, row)
		}

		return &QueryResult{
			Columns:       columns,
			Rows:          resultRows,
			RowsAffected:  int64(len(resultRows)),
			ExecutionTime: time.Since(start),
		}, rows.Err()
	} else {
		result, err := c.db.ExecContext(ctx, query)
		if err != nil {
			return &QueryResult{Error: err.Error()}, err
		}

		rowsAffected, _ := result.RowsAffected()

		return &QueryResult{
			RowsAffected:  rowsAffected,
			ExecutionTime: time.Since(start),
		}, nil
	}
}

func (c *PostgreSQLClient) InsertRow(ctx context.Context, tableName, schema string, data map[string]any) error {
	if schema == "" {
		schema = "public"
	}

	columns := []string{}
	placeholders := []string{}
	values := []any{}
	i := 1

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s.%s (%s) VALUES (%s)",
		schema, tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := c.db.ExecContext(ctx, query, values...)
	return err
}

func (c *PostgreSQLClient) UpdateRow(ctx context.Context, tableName, schema string, primaryKey map[string]any, data map[string]any) error {
	if schema == "" {
		schema = "public"
	}

	setClauses := []string{}
	whereClause := []string{}
	values := []any{}
	i := 1

	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}

	for col, val := range primaryKey {
		whereClause = append(whereClause, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf(
		"UPDATE %s.%s SET %s WHERE %s",
		schema, tableName,
		strings.Join(setClauses, ", "),
		strings.Join(whereClause, " AND "),
	)

	_, err := c.db.ExecContext(ctx, query, values...)
	return err
}

func (c *PostgreSQLClient) DeleteRow(ctx context.Context, tableName, schema string, primaryKey map[string]any) error {
	if schema == "" {
		schema = "public"
	}

	whereClauses := []string{}
	values := []any{}
	i := 1

	for col, val := range primaryKey {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf(
		"DELETE FROM %s.%s WHERE %s",
		schema, tableName,
		strings.Join(whereClauses, " AND "),
	)

	_, err := c.db.ExecContext(ctx, query, values...)
	return err
}

func (c *PostgreSQLClient) GetDatabaseVersion(ctx context.Context) (string, error) {
	var version string
	err := c.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	return version, err
}

func (c *PostgreSQLClient) GetDatabaseSize(ctx context.Context) (int64, error) {
	var size int64
	err := c.db.QueryRowContext(ctx, "SELECT pg_database_size(current_database())").Scan(&size)
	return size, err
}

func mapPostgresType(nativeType string) ColumnType {
	nativeType = strings.ToLower(nativeType)
	switch {
	case strings.Contains(nativeType, "char"), strings.Contains(nativeType, "text"):
		return ColumnTypeString
	case strings.Contains(nativeType, "int"), strings.Contains(nativeType, "serial"):
		return ColumnTypeInteger
	case strings.Contains(nativeType, "float"), strings.Contains(nativeType, "double"), strings.Contains(nativeType, "numeric"), strings.Contains(nativeType, "decimal"):
		return ColumnTypeFloat
	case nativeType == "boolean", nativeType == "bool":
		return ColumnTypeBoolean
	case nativeType == "date":
		return ColumnTypeDate
	case strings.Contains(nativeType, "timestamp"):
		return ColumnTypeTimestamp
	case nativeType == "json", nativeType == "jsonb":
		return ColumnTypeJSON
	case nativeType == "bytea":
		return ColumnTypeBinary
	case nativeType == "uuid":
		return ColumnTypeUUID
	case strings.HasSuffix(nativeType, "[]"):
		return ColumnTypeArray
	default:
		return ColumnTypeUnknown
	}
}

func buildWhereClause(filter Filter, argIndex int) (string, any) {
	placeholder := fmt.Sprintf("$%d", argIndex)

	switch filter.Operator {
	case "=", "equals":
		return fmt.Sprintf("%s = %s", filter.Column, placeholder), filter.Value
	case "!=", "not_equals":
		return fmt.Sprintf("%s != %s", filter.Column, placeholder), filter.Value
	case ">", "greater_than":
		return fmt.Sprintf("%s > %s", filter.Column, placeholder), filter.Value
	case ">=", "greater_or_equal":
		return fmt.Sprintf("%s >= %s", filter.Column, placeholder), filter.Value
	case "<", "less_than":
		return fmt.Sprintf("%s < %s", filter.Column, placeholder), filter.Value
	case "<=", "less_or_equal":
		return fmt.Sprintf("%s <= %s", filter.Column, placeholder), filter.Value
	case "like", "contains":
		return fmt.Sprintf("%s LIKE %s", filter.Column, placeholder), fmt.Sprintf("%%%v%%", filter.Value)
	case "starts_with":
		return fmt.Sprintf("%s LIKE %s", filter.Column, placeholder), fmt.Sprintf("%v%%", filter.Value)
	case "ends_with":
		return fmt.Sprintf("%s LIKE %s", filter.Column, placeholder), fmt.Sprintf("%%%v", filter.Value)
	case "is_null":
		return fmt.Sprintf("%s IS NULL", filter.Column), nil
	case "is_not_null":
		return fmt.Sprintf("%s IS NOT NULL", filter.Column), nil
	default:
		return fmt.Sprintf("%s = %s", filter.Column, placeholder), filter.Value
	}
}
