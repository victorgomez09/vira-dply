package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mikrocloud/mikrocloud/internal/domain/applications"
	applicationService "github.com/mikrocloud/mikrocloud/internal/domain/applications/service"
	"github.com/mikrocloud/mikrocloud/internal/domain/services"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

// TemplateRepository interface for service template persistence
type TemplateRepository interface {
	Save(template *services.ServiceTemplate) error
	FindByID(id services.TemplateID) (*services.ServiceTemplate, error)
	FindByName(name services.TemplateName) (*services.ServiceTemplate, error)
	FindByCategory(category services.TemplateCategory) ([]*services.ServiceTemplate, error)
	ListOfficial() ([]*services.ServiceTemplate, error)
	Update(template *services.ServiceTemplate) error
	Delete(id services.TemplateID) error
	List() ([]*services.ServiceTemplate, error)
}

// QuickDeployService provides a service for quick application deployment from templates
type QuickDeployService struct {
	templateRepo TemplateRepository
	appService   ApplicationService // interface for applications domain
}

// ApplicationService interface to avoid circular dependencies
type ApplicationService interface {
	CreateApplication(ctx context.Context, cmd applicationService.CreateApplicationCommand) (*applications.Application, error)
}

func NewQuickDeployService(templateRepo TemplateRepository, appService ApplicationService) *QuickDeployService {
	return &QuickDeployService{
		templateRepo: templateRepo,
		appService:   appService,
	}
}

// DeployTemplate creates an application from a template
func (s *QuickDeployService) DeployTemplate(ctx context.Context, templateID services.TemplateID, req services.DeploymentRequest) (*applications.Application, error) {
	// Find the template
	template, err := s.templateRepo.FindByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	// Create application from template
	app, err := template.CreateApplication(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create application from template: %w", err)
	}

	// Convert Application to CreateApplicationCommand for the service
	cmd := applicationService.CreateApplicationCommand{
		Name:             app.Name().String(),
		Description:      app.Description(),
		ProjectID:        app.ProjectID(),
		EnvironmentID:    app.EnvironmentID(),
		DeploymentSource: app.DeploymentSource(),
		BuildpackConfig:  app.Buildpack(),
		EnvVars:          app.EnvVars(),
	}

	// Create the application using the service
	createdApp, err := s.appService.CreateApplication(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to save application: %w", err)
	}

	return createdApp, nil
}

// SQLiteTemplateRepository implements the TemplateRepository interface
type SQLiteTemplateRepository struct {
	db *sql.DB
}

// NewSQLiteTemplateRepository creates a new SQLite-based template repository
func NewSQLiteTemplateRepository(db *sql.DB) TemplateRepository {
	return &SQLiteTemplateRepository{db: db}
}

// Save persists a service template to the database
func (r *SQLiteTemplateRepository) Save(template *services.ServiceTemplate) error {
	ctx := context.Background()

	// Marshal GitURL to JSON
	var gitURLJSON string
	if template.GitURL() != nil {
		gitURLData := map[string]string{
			"url":          template.GitURL().URL(),
			"branch":       template.GitURL().Branch(),
			"context_root": template.GitURL().ContextRoot(),
		}
		data, err := json.Marshal(gitURLData)
		if err != nil {
			return fmt.Errorf("failed to marshal git URL: %w", err)
		}
		gitURLJSON = string(data)
	}

	// Marshal BuildConfig to JSON
	var buildConfigJSON string
	if template.BuildConfig() != nil {
		buildConfigData := map[string]interface{}{
			"buildpack_type": string(template.BuildConfig().BuildpackType()),
		}
		if template.BuildConfig().NixpacksConfig() != nil {
			buildConfigData["nixpacks"] = template.BuildConfig().NixpacksConfig()
		}
		if template.BuildConfig().StaticConfig() != nil {
			buildConfigData["static"] = template.BuildConfig().StaticConfig()
		}
		if template.BuildConfig().DockerfileConfig() != nil {
			buildConfigData["dockerfile"] = template.BuildConfig().DockerfileConfig()
		}
		if template.BuildConfig().ComposeConfig() != nil {
			buildConfigData["compose"] = template.BuildConfig().ComposeConfig()
		}
		data, err := json.Marshal(buildConfigData)
		if err != nil {
			return fmt.Errorf("failed to marshal build config: %w", err)
		}
		buildConfigJSON = string(data)
	}

	// Marshal environment variables to JSON
	envJSON, err := json.Marshal(template.Environment())
	if err != nil {
		return fmt.Errorf("failed to marshal environment: %w", err)
	}

	// Marshal ports to JSON
	portsJSON, err := json.Marshal(template.Ports())
	if err != nil {
		return fmt.Errorf("failed to marshal ports: %w", err)
	}

	// Marshal volumes to JSON
	volumesJSON, err := json.Marshal(template.Volumes())
	if err != nil {
		return fmt.Errorf("failed to marshal volumes: %w", err)
	}

	// Use Bob query builder for INSERT with ON CONFLICT (upsert)
	query := sqlite.Insert(
		im.Into("service_templates"),
		im.Values(
			sqlite.Arg(template.ID().String()),
			sqlite.Arg(template.Name().String()),
			sqlite.Arg(template.Description()),
			sqlite.Arg(string(template.Category())),
			sqlite.Arg(template.Version()),
			sqlite.Arg(gitURLJSON),
			sqlite.Arg(buildConfigJSON),
			sqlite.Arg(string(envJSON)),
			sqlite.Arg(string(portsJSON)),
			sqlite.Arg(string(volumesJSON)),
			sqlite.Arg(template.IsOfficial()),
			sqlite.Arg(template.CreatedAt().Format(time.RFC3339)),
			sqlite.Arg(template.UpdatedAt().Format(time.RFC3339)),
		),
		im.OnConflict("id").DoUpdate(
			im.SetCol("description").ToArg(template.Description()),
			im.SetCol("version").ToArg(template.Version()),
			im.SetCol("git_url").ToArg(gitURLJSON),
			im.SetCol("build_config").ToArg(buildConfigJSON),
			im.SetCol("environment").ToArg(string(envJSON)),
			im.SetCol("ports").ToArg(string(portsJSON)),
			im.SetCol("volumes").ToArg(string(volumesJSON)),
			im.SetCol("updated_at").ToArg(template.UpdatedAt().Format(time.RFC3339)),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// FindByID retrieves a template by its ID
func (r *SQLiteTemplateRepository) FindByID(id services.TemplateID) (*services.ServiceTemplate, error) {
	ctx := context.Background()

	query := sqlite.Select(
		sm.Columns("id", "name", "description", "category", "version", "git_url", "build_config", "environment", "ports", "volumes", "is_official", "created_at", "updated_at"),
		sm.From("service_templates"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row templateRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Name, &row.Description, &row.Category, &row.Version,
		&row.GitURL, &row.BuildConfig, &row.Environment, &row.Ports, &row.Volumes,
		&row.IsOfficial, &row.CreatedAt, &row.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to find template by ID: %w", err)
	}

	return r.mapRowToTemplate(row)
}

// FindByName retrieves a template by its name
func (r *SQLiteTemplateRepository) FindByName(name services.TemplateName) (*services.ServiceTemplate, error) {
	ctx := context.Background()

	query := `
		SELECT id, name, description, category, version, git_url, build_config, environment, ports, volumes, is_official, created_at, updated_at
		FROM service_templates 
		WHERE name = ?`

	var row templateRow
	err := r.db.QueryRowContext(ctx, query, name.String()).Scan(
		&row.ID, &row.Name, &row.Description, &row.Category, &row.Version,
		&row.GitURL, &row.BuildConfig, &row.Environment, &row.Ports, &row.Volumes,
		&row.IsOfficial, &row.CreatedAt, &row.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found: %s", name.String())
		}
		return nil, fmt.Errorf("failed to find template by name: %w", err)
	}

	return r.mapRowToTemplate(row)
}

// FindByCategory retrieves all templates in a category
func (r *SQLiteTemplateRepository) FindByCategory(category services.TemplateCategory) ([]*services.ServiceTemplate, error) {
	ctx := context.Background()

	query := `
		SELECT id, name, description, category, version, git_url, build_config, environment, ports, volumes, is_official, created_at, updated_at
		FROM service_templates 
		WHERE category = ?
		ORDER BY is_official DESC, name ASC`

	return r.queryTemplates(ctx, query, string(category))
}

// ListOfficial retrieves all official templates
func (r *SQLiteTemplateRepository) ListOfficial() ([]*services.ServiceTemplate, error) {
	ctx := context.Background()

	query := `
		SELECT id, name, description, category, version, git_url, build_config, environment, ports, volumes, is_official, created_at, updated_at
		FROM service_templates 
		WHERE is_official = true
		ORDER BY category ASC, name ASC`

	return r.queryTemplates(ctx, query)
}

// Update updates an existing template
func (r *SQLiteTemplateRepository) Update(template *services.ServiceTemplate) error {
	// Implementation similar to Save but using UPDATE
	return r.Save(template) // For simplicity, using upsert
}

// Delete removes a template
func (r *SQLiteTemplateRepository) Delete(id services.TemplateID) error {
	ctx := context.Background()

	query := sqlite.Delete(
		dm.From("service_templates"),
		dm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %s", id.String())
	}

	return nil
}

// List retrieves all templates
func (r *SQLiteTemplateRepository) List() ([]*services.ServiceTemplate, error) {
	ctx := context.Background()

	query := `
		SELECT id, name, description, category, version, git_url, build_config, environment, ports, volumes, is_official, created_at, updated_at
		FROM service_templates 
		ORDER BY is_official DESC, category ASC, name ASC`

	return r.queryTemplates(ctx, query)
}

// Helper method to query templates
func (r *SQLiteTemplateRepository) queryTemplates(ctx context.Context, query string, args ...interface{}) ([]*services.ServiceTemplate, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []*services.ServiceTemplate
	for rows.Next() {
		var row templateRow
		err := rows.Scan(&row.ID, &row.Name, &row.Description, &row.Category, &row.Version,
			&row.GitURL, &row.BuildConfig, &row.Environment, &row.Ports, &row.Volumes,
			&row.IsOfficial, &row.CreatedAt, &row.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template row: %w", err)
		}

		template, err := r.mapRowToTemplate(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map template: %w", err)
		}

		templates = append(templates, template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over template rows: %w", err)
	}

	return templates, nil
}

// templateRow represents the database row structure
type templateRow struct {
	ID          string
	Name        string
	Description string
	Category    string
	Version     string
	GitURL      string
	BuildConfig string
	Environment string
	Ports       string
	Volumes     string
	IsOfficial  bool
	CreatedAt   string
	UpdatedAt   string
}

// mapRowToTemplate converts a database row to a domain ServiceTemplate
func (r *SQLiteTemplateRepository) mapRowToTemplate(row templateRow) (*services.ServiceTemplate, error) {
	// Parse template ID
	templateID, err := services.TemplateIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}

	// Parse template name
	templateName, err := services.NewTemplateName(row.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	// Parse category
	category := services.TemplateCategory(row.Category)

	// Parse GitURL
	var gitURL *applications.GitURL
	if row.GitURL != "" {
		var gitURLData map[string]string
		err = json.Unmarshal([]byte(row.GitURL), &gitURLData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal git URL: %w", err)
		}
		gitURL, err = applications.NewGitURL(
			gitURLData["url"],
			gitURLData["branch"],
			gitURLData["context_root"],
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create git URL: %w", err)
		}
	}

	// Parse BuildConfig
	var buildConfig *applications.BuildConfig
	if row.BuildConfig != "" {
		var buildConfigData map[string]interface{}
		err = json.Unmarshal([]byte(row.BuildConfig), &buildConfigData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal build config: %w", err)
		}

		buildpackType := applications.BuildpackType(buildConfigData["buildpack_type"].(string))
		buildConfig = applications.NewBuildConfig(buildpackType)

		// Set specific config based on buildpack type
		switch buildpackType {
		case applications.BuildpackTypeNixpacks:
			if nixpacks, exists := buildConfigData["nixpacks"]; exists {
				configBytes, _ := json.Marshal(nixpacks)
				var nixpacksConfig applications.NixpacksConfig
				if err := json.Unmarshal(configBytes, &nixpacksConfig); err == nil {
					buildConfig.SetNixpacksConfig(&nixpacksConfig)
				}
			}
		case applications.BuildpackTypeStatic:
			if static, exists := buildConfigData["static"]; exists {
				configBytes, _ := json.Marshal(static)
				var staticConfig applications.StaticConfig
				if err := json.Unmarshal(configBytes, &staticConfig); err == nil {
					buildConfig.SetStaticConfig(&staticConfig)
				}
			}
		case applications.BuildpackTypeDockerfile:
			if dockerfile, exists := buildConfigData["dockerfile"]; exists {
				configBytes, _ := json.Marshal(dockerfile)
				var dockerfileConfig applications.DockerfileConfig
				if err := json.Unmarshal(configBytes, &dockerfileConfig); err == nil {
					buildConfig.SetDockerfileConfig(&dockerfileConfig)
				}
			}
		case applications.BuildpackTypeDockerCompose:
			if compose, exists := buildConfigData["compose"]; exists {
				configBytes, _ := json.Marshal(compose)
				var composeConfig applications.ComposeConfig
				if err := json.Unmarshal(configBytes, &composeConfig); err == nil {
					buildConfig.SetComposeConfig(&composeConfig)
				}
			}
		}
	}

	// Parse environment variables
	var environment map[string]string
	if row.Environment != "" {
		err = json.Unmarshal([]byte(row.Environment), &environment)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal environment: %w", err)
		}
	} else {
		environment = make(map[string]string)
	}

	// Parse ports
	var ports []services.Port
	if row.Ports != "" {
		err = json.Unmarshal([]byte(row.Ports), &ports)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal ports: %w", err)
		}
	} else {
		ports = make([]services.Port, 0)
	}

	// Parse volumes
	var volumes []services.Volume
	if row.Volumes != "" {
		err = json.Unmarshal([]byte(row.Volumes), &volumes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal volumes: %w", err)
		}
	} else {
		volumes = make([]services.Volume, 0)
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
	}

	// Reconstruct template from persistence
	template := services.ReconstructServiceTemplate(
		templateID, templateName, row.Description, category, row.Version,
		gitURL, buildConfig, environment, ports, volumes, row.IsOfficial,
		createdAt, updatedAt)

	return template, nil
}
