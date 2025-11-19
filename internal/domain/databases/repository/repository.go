package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
)

type DatabaseRepository interface {
	Create(database *databases.Database) error
	GetByID(id databases.DatabaseID) (*databases.Database, error)
	GetByName(projectID uuid.UUID, name databases.DatabaseName) (*databases.Database, error)
	ListByProject(projectID uuid.UUID) ([]*databases.Database, error)
	ListByEnvironment(projectID, environmentID uuid.UUID) ([]*databases.Database, error)
	ListAllWithContainers() ([]*databases.Database, error)
	Update(database *databases.Database) error
	Delete(id databases.DatabaseID) error
	ExistsByName(projectID uuid.UUID, name databases.DatabaseName) (bool, error)
}

type SQLiteDatabaseRepository struct {
	db *sql.DB
}

func NewSQLiteDatabaseRepository(db *sql.DB) DatabaseRepository {
	return &SQLiteDatabaseRepository{db: db}
}

func (r *SQLiteDatabaseRepository) Create(database *databases.Database) error {
	configJSON, err := database.ConfigJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	portsJSON, err := json.Marshal(database.Ports())
	if err != nil {
		return fmt.Errorf("failed to marshal ports: %w", err)
	}

	query := `
		INSERT INTO databases (
			id, name, description, type, project_id, environment_id,
			config, status, connection_string, ports, container_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		database.ID().String(),
		database.Name().String(),
		database.Description(),
		string(database.Type()),
		database.ProjectID().String(),
		database.EnvironmentID().String(),
		string(configJSON),
		string(database.Status()),
		database.ConnectionString(),
		string(portsJSON),
		database.ContainerID(),
		database.CreatedAt(),
		database.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

func (r *SQLiteDatabaseRepository) GetByID(id databases.DatabaseID) (*databases.Database, error) {
	query := `
		SELECT id, name, description, type, project_id, environment_id,
			   config, status, connection_string, ports, container_id, created_at, updated_at
		FROM databases
		WHERE id = ?
	`

	row := r.db.QueryRow(query, id.String())
	return r.scanDatabase(row)
}

func (r *SQLiteDatabaseRepository) GetByName(projectID uuid.UUID, name databases.DatabaseName) (*databases.Database, error) {
	query := `
		SELECT id, name, description, type, project_id, environment_id,
			   config, status, connection_string, ports, container_id, created_at, updated_at
		FROM databases
		WHERE project_id = ? AND name = ?
	`

	row := r.db.QueryRow(query, projectID.String(), name.String())
	return r.scanDatabase(row)
}

func (r *SQLiteDatabaseRepository) ListByProject(projectID uuid.UUID) ([]*databases.Database, error) {
	query := `
		SELECT id, name, description, type, project_id, environment_id,
			   config, status, connection_string, ports, container_id, created_at, updated_at
		FROM databases
		WHERE project_id = ?
		ORDER BY name ASC
	`

	rows, err := r.db.Query(query, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list databases by project: %w", err)
	}
	defer rows.Close()

	return r.scanDatabases(rows)
}

func (r *SQLiteDatabaseRepository) ListByEnvironment(projectID, environmentID uuid.UUID) ([]*databases.Database, error) {
	query := `
		SELECT id, name, description, type, project_id, environment_id,
			   config, status, connection_string, ports, container_id, created_at, updated_at
		FROM databases
		WHERE project_id = ? AND environment_id = ?
		ORDER BY name ASC
	`

	rows, err := r.db.Query(query, projectID.String(), environmentID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list databases by environment: %w", err)
	}
	defer rows.Close()

	return r.scanDatabases(rows)
}

func (r *SQLiteDatabaseRepository) ListAllWithContainers() ([]*databases.Database, error) {
	query := `
		SELECT id, name, description, type, project_id, environment_id,
			   config, status, connection_string, ports, container_id, created_at, updated_at
		FROM databases
		WHERE container_id != '' AND container_id IS NOT NULL
		ORDER BY name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases with containers: %w", err)
	}
	defer rows.Close()

	return r.scanDatabases(rows)
}

func (r *SQLiteDatabaseRepository) Update(database *databases.Database) error {
	configJSON, err := database.ConfigJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	portsJSON, err := json.Marshal(database.Ports())
	if err != nil {
		return fmt.Errorf("failed to marshal ports: %w", err)
	}

	query := `
		UPDATE databases
		SET name = ?, description = ?, type = ?, config = ?, status = ?,
		    connection_string = ?, ports = ?, container_id = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(
		query,
		database.Name().String(),
		database.Description(),
		string(database.Type()),
		string(configJSON),
		string(database.Status()),
		database.ConnectionString(),
		string(portsJSON),
		database.ContainerID(),
		database.UpdatedAt(),
		database.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("database not found")
	}

	return nil
}

func (r *SQLiteDatabaseRepository) Delete(id databases.DatabaseID) error {
	query := `DELETE FROM databases WHERE id = ?`

	result, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("database not found")
	}

	return nil
}

func (r *SQLiteDatabaseRepository) ExistsByName(projectID uuid.UUID, name databases.DatabaseName) (bool, error) {
	query := `SELECT COUNT(*) FROM databases WHERE project_id = ? AND name = ?`

	var count int
	err := r.db.QueryRow(query, projectID.String(), name.String()).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check database existence: %w", err)
	}

	return count > 0, nil
}

func (r *SQLiteDatabaseRepository) scanDatabase(row *sql.Row) (*databases.Database, error) {
	var (
		id, name, description, dbType, projectID, environmentID string
		configJSON, status, connectionString, portsJSON         string
		containerID                                             string
		createdAt, updatedAt                                    string
	)

	err := row.Scan(
		&id, &name, &description, &dbType, &projectID, &environmentID,
		&configJSON, &status, &connectionString, &portsJSON, &containerID, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("database not found")
		}
		return nil, fmt.Errorf("failed to scan database: %w", err)
	}

	return r.buildDatabaseFromRow(
		id, name, description, dbType, projectID, environmentID,
		configJSON, status, connectionString, portsJSON, containerID, createdAt, updatedAt,
	)
}

func (r *SQLiteDatabaseRepository) scanDatabases(rows *sql.Rows) ([]*databases.Database, error) {
	var result []*databases.Database

	for rows.Next() {
		var (
			id, name, description, dbType, projectID, environmentID string
			configJSON, status, connectionString, portsJSON         string
			containerID                                             string
			createdAt, updatedAt                                    string
		)

		err := rows.Scan(
			&id, &name, &description, &dbType, &projectID, &environmentID,
			&configJSON, &status, &connectionString, &portsJSON, &containerID, &createdAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan database row: %w", err)
		}

		database, err := r.buildDatabaseFromRow(
			id, name, description, dbType, projectID, environmentID,
			configJSON, status, connectionString, portsJSON, containerID, createdAt, updatedAt,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, database)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating database rows: %w", err)
	}

	return result, nil
}

func (r *SQLiteDatabaseRepository) buildDatabaseFromRow(
	idStr, nameStr, description, dbTypeStr, projectIDStr, environmentIDStr,
	configJSON, statusStr, connectionString, portsJSON, containerID, createdAtStr, updatedAtStr string,
) (*databases.Database, error) {
	// Parse IDs
	databaseID, err := databases.DatabaseIDFromString(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid database ID: %w", err)
	}

	databaseName, err := databases.NewDatabaseName(nameStr)
	if err != nil {
		return nil, fmt.Errorf("invalid database name: %w", err)
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	environmentID, err := uuid.Parse(environmentIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid environment ID: %w", err)
	}

	// Parse config
	var config databases.DatabaseConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Parse ports
	var ports map[string]int
	if err := json.Unmarshal([]byte(portsJSON), &ports); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ports: %w", err)
	}

	// Parse timestamps
	createdAt, err := parseTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	updatedAt, err := parseTime(updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	// Reconstruct database
	return databases.ReconstructDatabase(
		databaseID,
		databaseName,
		description,
		databases.DatabaseType(dbTypeStr),
		projectID,
		environmentID,
		config,
		databases.DatabaseStatus(statusStr),
		connectionString,
		ports,
		containerID,
		createdAt,
		updatedAt,
	), nil
}

// parseTime parses time string in RFC3339 format
func parseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}
