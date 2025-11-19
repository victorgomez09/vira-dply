package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/environments"
)

// Repository defines the interface for environment persistence
type Repository interface {
	Save(ctx context.Context, env *environments.Environment) error
	FindByID(ctx context.Context, id environments.EnvironmentID) (*environments.Environment, error)
	FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*environments.Environment, error)
	FindAll(ctx context.Context) ([]*environments.Environment, error)
	Delete(ctx context.Context, id environments.EnvironmentID) error
}

// SQLiteEnvironmentRepository implements Repository using SQLite
type SQLiteEnvironmentRepository struct {
	db *sql.DB
}

// NewSQLiteEnvironmentRepository creates a new SQLite environment repository
func NewSQLiteEnvironmentRepository(db *sql.DB) Repository {
	return &SQLiteEnvironmentRepository{db: db}
}

// Save persists an environment to the database using raw SQL
func (r *SQLiteEnvironmentRepository) Save(ctx context.Context, env *environments.Environment) error {
	// Check if environment exists
	existingQuery := `SELECT id FROM environments WHERE id = ?`
	var existingID string
	err := r.db.QueryRowContext(ctx, existingQuery, env.ID().String()).Scan(&existingID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing environment: %w", err)
	}

	if err == sql.ErrNoRows {
		// Insert new environment
		query := `
			INSERT INTO environments (id, name, is_production, project_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		_, err = r.db.ExecContext(ctx, query,
			env.ID().String(),
			env.Name().String(),
			boolToInt(env.IsProduction()),
			env.ProjectID().String(),
			env.CreatedAt().Format(time.RFC3339),
			env.UpdatedAt().Format(time.RFC3339),
		)
	} else {
		// Update existing environment
		query := `
			UPDATE environments 
			SET name = ?, is_production = ?, updated_at = ?
			WHERE id = ?
		`
		_, err = r.db.ExecContext(ctx, query,
			env.Name().String(),
			boolToInt(env.IsProduction()),
			env.UpdatedAt().Format(time.RFC3339),
			env.ID().String(),
		)
	}

	if err != nil {
		return fmt.Errorf("failed to save environment: %w", err)
	}

	return nil
}

// FindByID retrieves an environment by its ID using raw SQL
func (r *SQLiteEnvironmentRepository) FindByID(ctx context.Context, id environments.EnvironmentID) (*environments.Environment, error) {
	query := `
		SELECT id, name, is_production, project_id, created_at, updated_at
		FROM environments
		WHERE id = ?
	`

	var row environmentRow
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&row.ID,
		&row.Name,
		&row.IsProduction,
		&row.ProjectID,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("environment not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to find environment by ID: %w", err)
	}

	return r.mapRowToEnvironment(row)
}

// FindByProjectID retrieves all environments for a specific project using raw SQL
func (r *SQLiteEnvironmentRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*environments.Environment, error) {
	query := `
		SELECT id, name, is_production, project_id, created_at, updated_at
		FROM environments
		WHERE project_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query environments by project: %w", err)
	}
	defer rows.Close()

	var envs []*environments.Environment
	for rows.Next() {
		var row environmentRow
		err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.IsProduction,
			&row.ProjectID,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan environment row: %w", err)
		}

		env, err := r.mapRowToEnvironment(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map environment: %w", err)
		}

		envs = append(envs, env)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environment rows: %w", err)
	}

	return envs, nil
}

// FindAll retrieves all environments using raw SQL
func (r *SQLiteEnvironmentRepository) FindAll(ctx context.Context) ([]*environments.Environment, error) {
	query := `
		SELECT id, name, is_production, project_id, created_at, updated_at
		FROM environments
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all environments: %w", err)
	}
	defer rows.Close()

	var envs []*environments.Environment
	for rows.Next() {
		var row environmentRow
		err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.IsProduction,
			&row.ProjectID,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan environment row: %w", err)
		}

		env, err := r.mapRowToEnvironment(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map environment: %w", err)
		}

		envs = append(envs, env)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environment rows: %w", err)
	}

	return envs, nil
}

// Delete removes an environment from the database using raw SQL
func (r *SQLiteEnvironmentRepository) Delete(ctx context.Context, id environments.EnvironmentID) error {
	query := `DELETE FROM environments WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("environment not found: %s", id.String())
	}

	return nil
}

// environmentRow represents the database row structure matching the schema
type environmentRow struct {
	ID           string
	Name         string
	IsProduction int
	ProjectID    string
	CreatedAt    string
	UpdatedAt    string
}

// mapRowToEnvironment converts a database row to a domain Environment
func (r *SQLiteEnvironmentRepository) mapRowToEnvironment(row environmentRow) (*environments.Environment, error) {
	// Parse environment ID
	envID, err := environments.EnvironmentIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid environment ID: %w", err)
	}

	// Parse environment name
	envName, err := environments.NewEnvironmentName(row.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid environment name: %w", err)
	}

	// Parse project ID
	projectID := uuid.MustParse(row.ProjectID)

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
	}

	// Convert integer to boolean
	isProduction := row.IsProduction == 1

	// Reconstruct environment from persistence (empty description and variables for now)
	env := environments.ReconstructEnvironment(
		envID, envName, isProduction, projectID, "", make(map[string]string), createdAt, updatedAt)

	return env, nil
}

// boolToInt converts a boolean to an integer for SQLite storage
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
