package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/applications"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type Repository interface {
	Save(ctx context.Context, app *applications.Application) error
	FindByID(ctx context.Context, id applications.ApplicationID) (*applications.Application, error)
	FindByName(ctx context.Context, projectID uuid.UUID, name applications.ApplicationName) (*applications.Application, error)
	FindByProject(ctx context.Context, projectID uuid.UUID) ([]*applications.Application, error)
	FindByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*applications.Application, error)
	FindAll(ctx context.Context) ([]*applications.Application, error)
	Delete(ctx context.Context, id applications.ApplicationID) error
	Exists(ctx context.Context, projectID uuid.UUID, name applications.ApplicationName) (bool, error)
}

type SQLiteApplicationRepository struct {
	db *sql.DB
}

func NewSQLiteApplicationRepository(db *sql.DB) *SQLiteApplicationRepository {
	return &SQLiteApplicationRepository{db: db}
}

func (r *SQLiteApplicationRepository) Save(ctx context.Context, app *applications.Application) error {
	buildpackJSON, err := json.Marshal(app.Buildpack())
	if err != nil {
		return fmt.Errorf("failed to marshal buildpack config: %w", err)
	}

	exposedPortsJSON, err := json.Marshal(app.ExposedPorts())
	if err != nil {
		return fmt.Errorf("failed to marshal exposed ports: %w", err)
	}

	portMappingsJSON, err := json.Marshal(app.PortMappings())
	if err != nil {
		return fmt.Errorf("failed to marshal port mappings: %w", err)
	}

	query := sqlite.Insert(
		im.Into("applications"),
		im.Values(
			sqlite.Arg(app.ID().String()),
			sqlite.Arg(app.Name().String()),
			sqlite.Arg(app.Description()),
			sqlite.Arg(app.ProjectID().String()),
			sqlite.Arg(app.EnvironmentID().String()),
			sqlite.Arg(app.RepoURL()),
			sqlite.Arg(app.RepoBranch()),
			sqlite.Arg(app.RepoPath()),
			sqlite.Arg(app.Domain()),
			sqlite.Arg(string(app.BuildpackType())),
			sqlite.Arg(string(buildpackJSON)),
			sqlite.Arg(app.AutoDeploy()),
			sqlite.Arg(string(app.Status())),
			sqlite.Arg(app.CreatedAt().Format(time.RFC3339)),
			sqlite.Arg(app.UpdatedAt().Format(time.RFC3339)),
			sqlite.Arg(app.BasePath()),
			sqlite.Arg(app.GeneratedDomain()),
			sqlite.Arg(string(exposedPortsJSON)),
			sqlite.Arg(string(portMappingsJSON)),
		),
		im.OnConflict("id").DoUpdate(
			im.SetCol("name").ToArg(app.Name().String()),
			im.SetCol("description").ToArg(app.Description()),
			im.SetCol("repo_url").ToArg(app.RepoURL()),
			im.SetCol("repo_branch").ToArg(app.RepoBranch()),
			im.SetCol("repo_path").ToArg(app.RepoPath()),
			im.SetCol("domain").ToArg(app.Domain()),
			im.SetCol("buildpack_type").ToArg(string(app.BuildpackType())),
			im.SetCol("config").ToArg(string(buildpackJSON)),
			im.SetCol("auto_deploy").ToArg(app.AutoDeploy()),
			im.SetCol("status").ToArg(string(app.Status())),
			im.SetCol("updated_at").ToArg(app.UpdatedAt().Format(time.RFC3339)),
			im.SetCol("base_path").ToArg(app.BasePath()),
			im.SetCol("generated_domain").ToArg(app.GeneratedDomain()),
			im.SetCol("exposed_ports").ToArg(string(exposedPortsJSON)),
			im.SetCol("port_mappings").ToArg(string(portMappingsJSON)),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to save application: %w", err)
	}

	return nil
}

func (r *SQLiteApplicationRepository) FindByID(ctx context.Context, id applications.ApplicationID) (*applications.Application, error) {
	query := sqlite.Select(
		sm.Columns("id", "name", "description", "project_id", "environment_id", "repo_url", "repo_branch", "repo_path", "domain", "buildpack_type", "config", "auto_deploy", "status", "created_at", "updated_at", "base_path", "generated_domain", "exposed_ports", "port_mappings"),
		sm.From("applications"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row applicationRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Name, &row.Description, &row.ProjectID, &row.EnvironmentID,
		&row.RepoURL, &row.RepoBranch, &row.RepoPath, &row.Domain, &row.BuildpackType,
		&row.Config, &row.AutoDeploy, &row.Status, &row.CreatedAt, &row.UpdatedAt, &row.BasePath,
		&row.GeneratedDomain, &row.ExposedPorts, &row.PortMappings)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to find application by ID: %w", err)
	}

	return r.mapRowToApplication(row)
}

func (r *SQLiteApplicationRepository) FindByName(ctx context.Context, projectID uuid.UUID, name applications.ApplicationName) (*applications.Application, error) {
	query := sqlite.Select(
		sm.Columns("id", "name", "description", "project_id", "environment_id", "repo_url", "repo_branch", "repo_path", "domain", "buildpack_type", "config", "auto_deploy", "status", "created_at", "updated_at", "base_path", "generated_domain", "exposed_ports", "port_mappings"),
		sm.From("applications"),
		sm.Where(
			sqlite.Quote("project_id").EQ(sqlite.Arg(projectID.String())).
				And(sqlite.Quote("name").EQ(sqlite.Arg(name.String()))),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row applicationRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Name, &row.Description, &row.ProjectID, &row.EnvironmentID,
		&row.RepoURL, &row.RepoBranch, &row.RepoPath, &row.Domain, &row.BuildpackType,
		&row.Config, &row.AutoDeploy, &row.Status, &row.CreatedAt, &row.UpdatedAt, &row.BasePath,
		&row.GeneratedDomain, &row.ExposedPorts, &row.PortMappings)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found: %s in project %s", name.String(), projectID.String())
		}
		return nil, fmt.Errorf("failed to find application by name: %w", err)
	}

	return r.mapRowToApplication(row)
}

func (r *SQLiteApplicationRepository) FindByProject(ctx context.Context, projectID uuid.UUID) ([]*applications.Application, error) {
	query := sqlite.Select(
		sm.Columns("id", "name", "description", "project_id", "environment_id", "repo_url", "repo_branch", "repo_path", "domain", "buildpack_type", "config", "auto_deploy", "status", "created_at", "updated_at", "base_path", "generated_domain", "exposed_ports", "port_mappings"),
		sm.From("applications"),
		sm.Where(sqlite.Quote("project_id").EQ(sqlite.Arg(projectID.String()))),
		sm.OrderBy("created_at").Desc(),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications by project: %w", err)
	}
	defer rows.Close()

	var applications []*applications.Application
	for rows.Next() {
		var row applicationRow
		err := rows.Scan(&row.ID, &row.Name, &row.Description, &row.ProjectID, &row.EnvironmentID,
			&row.RepoURL, &row.RepoBranch, &row.RepoPath, &row.Domain, &row.BuildpackType,
			&row.Config, &row.AutoDeploy, &row.Status, &row.CreatedAt, &row.UpdatedAt, &row.BasePath,
			&row.GeneratedDomain, &row.ExposedPorts, &row.PortMappings)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application row: %w", err)
		}

		app, err := r.mapRowToApplication(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map application: %w", err)
		}

		applications = append(applications, app)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over application rows: %w", err)
	}

	return applications, nil
}

func (r *SQLiteApplicationRepository) FindByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*applications.Application, error) {
	query := sqlite.Select(
		sm.Columns("id", "name", "description", "project_id", "environment_id", "repo_url", "repo_branch", "repo_path", "domain", "buildpack_type", "config", "auto_deploy", "status", "created_at", "updated_at", "base_path", "generated_domain", "exposed_ports", "port_mappings"),
		sm.From("applications"),
		sm.Where(sqlite.Quote("environment_id").EQ(sqlite.Arg(environmentID.String()))),
		sm.OrderBy("created_at").Desc(),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications by environment: %w", err)
	}
	defer rows.Close()

	var applications []*applications.Application
	for rows.Next() {
		var row applicationRow
		err := rows.Scan(&row.ID, &row.Name, &row.Description, &row.ProjectID, &row.EnvironmentID,
			&row.RepoURL, &row.RepoBranch, &row.RepoPath, &row.Domain, &row.BuildpackType,
			&row.Config, &row.AutoDeploy, &row.Status, &row.CreatedAt, &row.UpdatedAt, &row.BasePath,
			&row.GeneratedDomain, &row.ExposedPorts, &row.PortMappings)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application row: %w", err)
		}

		app, err := r.mapRowToApplication(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map application: %w", err)
		}

		applications = append(applications, app)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over application rows: %w", err)
	}

	return applications, nil
}

func (r *SQLiteApplicationRepository) FindAll(ctx context.Context) ([]*applications.Application, error) {
	query := sqlite.Select(
		sm.Columns("id", "name", "description", "project_id", "environment_id", "repo_url", "repo_branch", "repo_path", "domain", "buildpack_type", "config", "auto_deploy", "status", "created_at", "updated_at", "base_path", "generated_domain", "exposed_ports", "port_mappings"),
		sm.From("applications"),
		sm.OrderBy("created_at").Desc(),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all applications: %w", err)
	}
	defer rows.Close()

	var applications []*applications.Application
	for rows.Next() {
		var row applicationRow
		err := rows.Scan(&row.ID, &row.Name, &row.Description, &row.ProjectID, &row.EnvironmentID,
			&row.RepoURL, &row.RepoBranch, &row.RepoPath, &row.Domain, &row.BuildpackType,
			&row.Config, &row.AutoDeploy, &row.Status, &row.CreatedAt, &row.UpdatedAt, &row.BasePath,
			&row.GeneratedDomain, &row.ExposedPorts, &row.PortMappings)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application row: %w", err)
		}

		app, err := r.mapRowToApplication(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map application: %w", err)
		}

		applications = append(applications, app)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over application rows: %w", err)
	}

	return applications, nil
}

func (r *SQLiteApplicationRepository) Delete(ctx context.Context, id applications.ApplicationID) error {
	query := sqlite.Delete(
		dm.From("applications"),
		dm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("application not found: %s", id.String())
	}

	return nil
}

func (r *SQLiteApplicationRepository) Exists(ctx context.Context, projectID uuid.UUID, name applications.ApplicationName) (bool, error) {
	query := sqlite.Select(
		sm.Columns("COUNT(*)"),
		sm.From("applications"),
		sm.Where(
			sqlite.Quote("project_id").EQ(sqlite.Arg(projectID.String())).
				And(sqlite.Quote("name").EQ(sqlite.Arg(name.String()))),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var count int
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if application exists: %w", err)
	}

	return count > 0, nil
}

type applicationRow struct {
	ID              string
	Name            string
	Description     sql.NullString
	ProjectID       string
	EnvironmentID   string
	RepoURL         sql.NullString
	RepoBranch      sql.NullString
	RepoPath        sql.NullString
	Domain          sql.NullString
	BuildpackType   string
	Config          string
	AutoDeploy      bool
	Status          string
	CreatedAt       string
	UpdatedAt       string
	BasePath        sql.NullString
	GeneratedDomain sql.NullString
	ExposedPorts    string
	PortMappings    string
}

func (r *SQLiteApplicationRepository) mapRowToApplication(row applicationRow) (*applications.Application, error) {
	appID, err := applications.ApplicationIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid application ID: %w", err)
	}

	appName, err := applications.NewApplicationName(row.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid application name: %w", err)
	}

	projectID := uuid.MustParse(row.ProjectID)
	environmentID := uuid.MustParse(row.EnvironmentID)

	// Parse build status and times
	status := applications.ApplicationStatus(row.Status)

	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
	}

	description := ""
	if row.Description.Valid {
		description = row.Description.String
	}

	domain := ""
	if row.Domain.Valid {
		domain = row.Domain.String
	}

	// Create deployment source from legacy fields for backward compatibility
	repoURL := ""
	if row.RepoURL.Valid {
		repoURL = row.RepoURL.String
	}

	repoBranch := "main"
	if row.RepoBranch.Valid {
		repoBranch = row.RepoBranch.String
	}

	repoPath := ""
	if row.RepoPath.Valid {
		repoPath = row.RepoPath.String
	}

	basePath := ""
	if row.BasePath.Valid {
		basePath = row.BasePath.String
	}

	// For now, create a git deployment source if we have a repo URL
	var deploymentSource applications.DeploymentSource
	if repoURL != "" {
		deploymentSource = applications.NewGitDeploymentSource(repoURL, repoBranch, repoPath, basePath)
	} else {
		// Default empty deployment source
		deploymentSource = applications.DeploymentSource{
			Type: applications.DeploymentSourceTypeGit,
		}
	}

	// Parse buildpack config from config field
	buildpackType := applications.BuildpackType(row.BuildpackType)
	var buildConfig *applications.BuildConfig

	// Try to parse config as legacy buildpack config format
	var parsedConfig applications.BuildpackConfig
	if err := json.Unmarshal([]byte(row.Config), &parsedConfig); err == nil && parsedConfig.Type != "" {
		// Convert legacy format to new format
		buildConfig = applications.NewLegacyBuildpackConfig(parsedConfig.Type, parsedConfig.Config)
	} else {
		// Fallback to simple config
		buildConfig = applications.NewBuildConfig(buildpackType)
	}

	envVars := make(map[string]string)

	generatedDomain := ""
	if row.GeneratedDomain.Valid {
		generatedDomain = row.GeneratedDomain.String
	}

	var exposedPorts []int
	if row.ExposedPorts != "" && row.ExposedPorts != "[]" {
		if err := json.Unmarshal([]byte(row.ExposedPorts), &exposedPorts); err != nil {
			exposedPorts = []int{}
		}
	}

	var portMappings []applications.PortMapping
	if row.PortMappings != "" && row.PortMappings != "[]" {
		if err := json.Unmarshal([]byte(row.PortMappings), &portMappings); err != nil {
			portMappings = []applications.PortMapping{}
		}
	}

	return applications.ReconstructApplication(
		appID, appName, description, projectID, environmentID,
		deploymentSource, domain, generatedDomain, exposedPorts, portMappings,
		buildConfig, envVars, row.AutoDeploy, status, createdAt, updatedAt), nil
}
