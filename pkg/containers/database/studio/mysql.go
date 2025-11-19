package studio

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLClient struct {
	db *sql.DB
}

func NewMySQLClient() *MySQLClient {
	return &MySQLClient{}
}

func (c *MySQLClient) Connect(ctx context.Context, connectionString string) error {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.db = db
	return nil
}

func (c *MySQLClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *MySQLClient) ListDatabases(ctx context.Context) ([]string, error) {
	query := `SHOW DATABASES`
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

func (c *MySQLClient) ListSchemas(ctx context.Context) ([]string, error) {
	return c.ListDatabases(ctx)
}

func (c *MySQLClient) ListTables(ctx context.Context, schema string) ([]string, error) {
	var query string
	var args []any

	if schema != "" {
		query = `SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME`
		args = []any{schema}
	} else {
		query = `SHOW TABLES`
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
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

func (c *MySQLClient) GetTableSchema(ctx context.Context, tableName, schema string) (*TableSchema, error) {
	if schema == "" {
		err := c.db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&schema)
		if err != nil {
			return nil, err
		}
	}

	tableSchema := &TableSchema{
		Name:   tableName,
		Schema: schema,
		Type:   "table",
	}

	columnsQuery := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			CHARACTER_MAXIMUM_LENGTH,
			COLUMN_KEY,
			EXTRA
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
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
		var columnKey, extra string
		var nativeType string
		var nullable string

		if err := rows.Scan(&col.Name, &nativeType, &nullable, &defaultVal, &maxLength, &columnKey, &extra); err != nil {
			return nil, err
		}

		col.NativeType = nativeType
		col.Type = mapMySQLType(nativeType)
		col.Nullable = nullable == "YES"

		if defaultVal.Valid {
			col.DefaultValue = &defaultVal.String
		}

		if maxLength.Valid {
			maxLen := int(maxLength.Int64)
			col.MaxLength = &maxLen
		}

		col.IsPrimaryKey = columnKey == "PRI"
		col.IsUnique = columnKey == "UNI" || col.IsPrimaryKey
		col.IsAutoInc = strings.Contains(extra, "auto_increment")

		if col.IsPrimaryKey {
			tableSchema.PrimaryKeys = append(tableSchema.PrimaryKeys, col.Name)
		}

		tableSchema.Columns = append(tableSchema.Columns, col)
	}

	fkQuery := `
		SELECT 
			CONSTRAINT_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ? 
			AND TABLE_NAME = ?
			AND REFERENCED_TABLE_NAME IS NOT NULL
	`

	fkRows, err := c.db.QueryContext(ctx, fkQuery, schema, tableName)
	if err != nil {
		return nil, err
	}
	defer fkRows.Close()

	for fkRows.Next() {
		var fk ForeignKey
		if err := fkRows.Scan(&fk.Name, &fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn); err != nil {
			return nil, err
		}
		tableSchema.ForeignKeys = append(tableSchema.ForeignKeys, fk)
	}

	indexQuery := `
		SELECT 
			INDEX_NAME,
			COLUMN_NAME,
			NON_UNIQUE,
			INDEX_TYPE
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`

	indexRows, err := c.db.QueryContext(ctx, indexQuery, schema, tableName)
	if err != nil {
		return nil, err
	}
	defer indexRows.Close()

	indexMap := make(map[string]*Index)
	for indexRows.Next() {
		var indexName, columnName, indexType string
		var nonUnique int

		if err := indexRows.Scan(&indexName, &columnName, &nonUnique, &indexType); err != nil {
			return nil, err
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &Index{
				Name:     indexName,
				Columns:  []string{columnName},
				IsUnique: nonUnique == 0,
				Type:     indexType,
			}
		}
	}

	for _, idx := range indexMap {
		tableSchema.Indexes = append(tableSchema.Indexes, *idx)
	}

	var rowCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", schema, tableName)
	if err := c.db.QueryRowContext(ctx, countQuery).Scan(&rowCount); err == nil {
		tableSchema.RowCount = &rowCount
	}

	return tableSchema, nil
}

func (c *MySQLClient) GetTableData(ctx context.Context, tableName, schema string, opts TableDataOptions) (*QueryResult, error) {
	if schema == "" {
		err := c.db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&schema)
		if err != nil {
			return nil, err
		}
	}

	start := time.Now()

	query := fmt.Sprintf("SELECT * FROM `%s`.`%s`", schema, tableName)

	whereClauses := []string{}
	args := []any{}

	for _, filter := range opts.Filters {
		clause, arg := buildMySQLWhereClause(filter)
		whereClauses = append(whereClauses, clause)
		if arg != nil {
			args = append(args, arg)
		}
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
			orderParts = append(orderParts, fmt.Sprintf("`%s` %s", sort.Column, direction))
		}
		query += " ORDER BY " + strings.Join(orderParts, ", ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", schema, tableName)
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

func (c *MySQLClient) ExecuteQuery(ctx context.Context, query string, opts QueryOptions) (*QueryResult, error) {
	start := time.Now()

	trimmedQuery := strings.TrimSpace(strings.ToUpper(query))
	isSelect := strings.HasPrefix(trimmedQuery, "SELECT") || strings.HasPrefix(trimmedQuery, "SHOW") || strings.HasPrefix(trimmedQuery, "DESCRIBE")

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

func (c *MySQLClient) InsertRow(ctx context.Context, tableName, schema string, data map[string]any) error {
	if schema == "" {
		err := c.db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&schema)
		if err != nil {
			return err
		}
	}

	columns := []string{}
	placeholders := []string{}
	values := []any{}

	for col, val := range data {
		columns = append(columns, fmt.Sprintf("`%s`", col))
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s`.`%s` (%s) VALUES (%s)",
		schema, tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := c.db.ExecContext(ctx, query, values...)
	return err
}

func (c *MySQLClient) UpdateRow(ctx context.Context, tableName, schema string, primaryKey map[string]any, data map[string]any) error {
	if schema == "" {
		err := c.db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&schema)
		if err != nil {
			return err
		}
	}

	setClauses := []string{}
	whereClauses := []string{}
	values := []any{}

	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", col))
		values = append(values, val)
	}

	for col, val := range primaryKey {
		whereClauses = append(whereClauses, fmt.Sprintf("`%s` = ?", col))
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"UPDATE `%s`.`%s` SET %s WHERE %s",
		schema, tableName,
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "),
	)

	_, err := c.db.ExecContext(ctx, query, values...)
	return err
}

func (c *MySQLClient) DeleteRow(ctx context.Context, tableName, schema string, primaryKey map[string]any) error {
	if schema == "" {
		err := c.db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&schema)
		if err != nil {
			return err
		}
	}

	whereClauses := []string{}
	values := []any{}

	for col, val := range primaryKey {
		whereClauses = append(whereClauses, fmt.Sprintf("`%s` = ?", col))
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"DELETE FROM `%s`.`%s` WHERE %s",
		schema, tableName,
		strings.Join(whereClauses, " AND "),
	)

	_, err := c.db.ExecContext(ctx, query, values...)
	return err
}

func (c *MySQLClient) GetDatabaseVersion(ctx context.Context) (string, error) {
	var version string
	err := c.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	return version, err
}

func (c *MySQLClient) GetDatabaseSize(ctx context.Context) (int64, error) {
	var size sql.NullInt64
	query := `
		SELECT SUM(data_length + index_length) 
		FROM information_schema.TABLES 
		WHERE table_schema = DATABASE()
	`
	err := c.db.QueryRowContext(ctx, query).Scan(&size)
	if !size.Valid {
		return 0, err
	}
	return size.Int64, err
}

func mapMySQLType(nativeType string) ColumnType {
	nativeType = strings.ToLower(nativeType)
	switch {
	case strings.Contains(nativeType, "char"), strings.Contains(nativeType, "text"):
		return ColumnTypeString
	case strings.Contains(nativeType, "int"), strings.Contains(nativeType, "serial"):
		return ColumnTypeInteger
	case strings.Contains(nativeType, "float"), strings.Contains(nativeType, "double"), strings.Contains(nativeType, "decimal"):
		return ColumnTypeFloat
	case nativeType == "boolean", nativeType == "bool", nativeType == "tinyint(1)":
		return ColumnTypeBoolean
	case nativeType == "date":
		return ColumnTypeDate
	case strings.Contains(nativeType, "datetime"):
		return ColumnTypeDateTime
	case strings.Contains(nativeType, "timestamp"):
		return ColumnTypeTimestamp
	case nativeType == "json":
		return ColumnTypeJSON
	case strings.Contains(nativeType, "blob"), strings.Contains(nativeType, "binary"):
		return ColumnTypeBinary
	default:
		return ColumnTypeUnknown
	}
}

func buildMySQLWhereClause(filter Filter) (string, any) {
	switch filter.Operator {
	case "=", "equals":
		return fmt.Sprintf("`%s` = ?", filter.Column), filter.Value
	case "!=", "not_equals":
		return fmt.Sprintf("`%s` != ?", filter.Column), filter.Value
	case ">", "greater_than":
		return fmt.Sprintf("`%s` > ?", filter.Column), filter.Value
	case ">=", "greater_or_equal":
		return fmt.Sprintf("`%s` >= ?", filter.Column), filter.Value
	case "<", "less_than":
		return fmt.Sprintf("`%s` < ?", filter.Column), filter.Value
	case "<=", "less_or_equal":
		return fmt.Sprintf("`%s` <= ?", filter.Column), filter.Value
	case "like", "contains":
		return fmt.Sprintf("`%s` LIKE ?", filter.Column), fmt.Sprintf("%%%v%%", filter.Value)
	case "starts_with":
		return fmt.Sprintf("`%s` LIKE ?", filter.Column), fmt.Sprintf("%v%%", filter.Value)
	case "ends_with":
		return fmt.Sprintf("`%s` LIKE ?", filter.Column), fmt.Sprintf("%%%v", filter.Value)
	case "is_null":
		return fmt.Sprintf("`%s` IS NULL", filter.Column), nil
	case "is_not_null":
		return fmt.Sprintf("`%s` IS NOT NULL", filter.Column), nil
	default:
		return fmt.Sprintf("`%s` = ?", filter.Column), filter.Value
	}
}
