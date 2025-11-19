// Package services contains deployment templates and marketplace functionality
package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/applications"
)

// ServiceTemplate represents a pre-configured deployment template
type ServiceTemplate struct {
	id          TemplateID
	name        TemplateName
	description string
	category    TemplateCategory
	version     string
	gitURL      *applications.GitURL
	buildConfig *applications.BuildConfig
	environment map[string]string
	ports       []Port
	volumes     []Volume
	isOfficial  bool
	createdAt   time.Time
	updatedAt   time.Time
}

// TemplateID is a value object for template identification
type TemplateID struct {
	value uuid.UUID
}

func NewTemplateID() TemplateID {
	return TemplateID{value: uuid.New()}
}

func TemplateIDFromString(s string) (TemplateID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TemplateID{}, fmt.Errorf("invalid template ID: %w", err)
	}
	return TemplateID{value: id}, nil
}

func (id TemplateID) String() string {
	return id.value.String()
}

func (id TemplateID) UUID() uuid.UUID {
	return id.value
}

// TemplateName is a value object that enforces naming rules
type TemplateName struct {
	value string
}

func NewTemplateName(name string) (TemplateName, error) {
	if name == "" {
		return TemplateName{}, fmt.Errorf("template name cannot be empty")
	}

	if len(name) > 50 {
		return TemplateName{}, fmt.Errorf("template name cannot exceed 50 characters")
	}

	return TemplateName{value: name}, nil
}

func (n TemplateName) String() string {
	return n.value
}

// TemplateCategory represents different types of service templates
type TemplateCategory string

const (
	CategoryDatabase   TemplateCategory = "database"
	CategoryWebApp     TemplateCategory = "webapp"
	CategoryAPI        TemplateCategory = "api"
	CategoryWorker     TemplateCategory = "worker"
	CategoryStorage    TemplateCategory = "storage"
	CategoryMonitoring TemplateCategory = "monitoring"
	CategoryCache      TemplateCategory = "cache"
	CategoryMessaging  TemplateCategory = "messaging"
	CategoryAnalytics  TemplateCategory = "analytics"
	CategorySecurity   TemplateCategory = "security"
	CategoryDevTools   TemplateCategory = "devtools"
	CategoryOther      TemplateCategory = "other"
)

// Port represents a port configuration
type Port struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // tcp, udp
	Public   bool   `json:"public"`   // whether to expose publicly
}

// Volume represents a volume mount
type Volume struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`
	Size      string `json:"size,omitempty"` // e.g., "1Gi", "500Mi"
	ReadOnly  bool   `json:"read_only"`
}

// DeploymentRequest represents a request to deploy a template as an application
type DeploymentRequest struct {
	ProjectID     uuid.UUID
	EnvironmentID uuid.UUID
	Name          string
	Environment   map[string]string      // override environment variables
	CustomConfig  map[string]interface{} // template-specific configuration
}

// DeploymentPreview shows what an application would look like from a template
type DeploymentPreview struct {
	TemplateName    TemplateName
	ApplicationName string
	ProjectID       uuid.UUID
	EnvironmentID   uuid.UUID
	GitURL          *applications.GitURL
	BuildConfig     *applications.BuildConfig
	Environment     map[string]string
	Ports           []Port
	Volumes         []Volume
}

// NewServiceTemplate creates a new service template
func NewServiceTemplate(
	name TemplateName,
	description string,
	category TemplateCategory,
	version string,
	gitURL *applications.GitURL,
	buildConfig *applications.BuildConfig,
	isOfficial bool,
) *ServiceTemplate {
	now := time.Now()

	return &ServiceTemplate{
		id:          NewTemplateID(),
		name:        name,
		description: description,
		category:    category,
		version:     version,
		gitURL:      gitURL,
		buildConfig: buildConfig,
		environment: make(map[string]string),
		ports:       make([]Port, 0),
		volumes:     make([]Volume, 0),
		isOfficial:  isOfficial,
		createdAt:   now,
		updatedAt:   now,
	}
}

// Getters
func (st *ServiceTemplate) ID() TemplateID {
	return st.id
}

func (st *ServiceTemplate) Name() TemplateName {
	return st.name
}

func (st *ServiceTemplate) Description() string {
	return st.description
}

func (st *ServiceTemplate) Category() TemplateCategory {
	return st.category
}

func (st *ServiceTemplate) Version() string {
	return st.version
}

func (st *ServiceTemplate) GitURL() *applications.GitURL {
	return st.gitURL
}

func (st *ServiceTemplate) BuildConfig() *applications.BuildConfig {
	return st.buildConfig
}

func (st *ServiceTemplate) Environment() map[string]string {
	// Return a copy to maintain encapsulation
	env := make(map[string]string)
	for k, v := range st.environment {
		env[k] = v
	}
	return env
}

func (st *ServiceTemplate) Ports() []Port {
	return append([]Port(nil), st.ports...)
}

func (st *ServiceTemplate) Volumes() []Volume {
	return append([]Volume(nil), st.volumes...)
}

func (st *ServiceTemplate) IsOfficial() bool {
	return st.isOfficial
}

func (st *ServiceTemplate) CreatedAt() time.Time {
	return st.createdAt
}

func (st *ServiceTemplate) UpdatedAt() time.Time {
	return st.updatedAt
}

// Business methods
func (st *ServiceTemplate) SetEnvironmentVariable(key, value string) error {
	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}

	st.environment[key] = value
	st.updatedAt = time.Now()
	return nil
}

func (st *ServiceTemplate) RemoveEnvironmentVariable(key string) {
	delete(st.environment, key)
	st.updatedAt = time.Now()
}

func (st *ServiceTemplate) AddPort(port Port) {
	st.ports = append(st.ports, port)
	st.updatedAt = time.Now()
}

func (st *ServiceTemplate) AddVolume(volume Volume) {
	st.volumes = append(st.volumes, volume)
	st.updatedAt = time.Now()
}

func (st *ServiceTemplate) UpdateVersion(version string) {
	st.version = version
	st.updatedAt = time.Now()
}

// CreateApplication creates an application from this template
func (st *ServiceTemplate) CreateApplication(req DeploymentRequest) (*applications.Application, error) {
	// Create application name
	appName, err := applications.NewApplicationName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid application name: %w", err)
	}

	// Create deployment source based on template configuration
	var deploymentSource applications.DeploymentSource
	if st.gitURL != nil {
		// Git-based deployment
		deploymentSource = applications.DeploymentSource{
			Type: applications.DeploymentSourceTypeGit,
			GitRepo: &applications.GitRepoSource{
				URL:    st.gitURL.URL(),
				Branch: st.gitURL.Branch(),
				Path:   st.gitURL.ContextRoot(),
			},
		}
	} else {
		// Docker registry deployment (for pre-built images like PostgreSQL)
		// For templates without Git URL, assume it's a Docker image
		imageName := st.name.String()
		imageTag := st.version
		if st.category == CategoryDatabase && st.name.String() == "postgres" {
			imageName = "postgres"
		}
		deploymentSource = applications.DeploymentSource{
			Type: applications.DeploymentSourceTypeRegistry,
			Registry: &applications.RegistrySource{
				Image: imageName,
				Tag:   imageTag,
			},
		}
	}

	// Create application from template
	app := applications.NewApplication(
		appName,
		st.description,
		req.ProjectID,
		req.EnvironmentID,
		deploymentSource,
		st.buildConfig,
	)

	// Merge template environment with request overrides
	finalEnv := make(map[string]string)
	for k, v := range st.environment {
		finalEnv[k] = v
	}
	for k, v := range req.Environment {
		finalEnv[k] = v
	}

	// Set environment variables
	for key, value := range finalEnv {
		app.SetEnvVar(key, value)
	}

	return app, nil
}

// ReconstructServiceTemplate recreates a service template from persistence data
func ReconstructServiceTemplate(
	id TemplateID,
	name TemplateName,
	description string,
	category TemplateCategory,
	version string,
	gitURL *applications.GitURL,
	buildConfig *applications.BuildConfig,
	environment map[string]string,
	ports []Port,
	volumes []Volume,
	isOfficial bool,
	createdAt time.Time,
	updatedAt time.Time,
) *ServiceTemplate {
	return &ServiceTemplate{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		version:     version,
		gitURL:      gitURL,
		buildConfig: buildConfig,
		environment: environment,
		ports:       ports,
		volumes:     volumes,
		isOfficial:  isOfficial,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// GetOfficialTemplates returns a list of official service templates
func GetOfficialTemplates() []*ServiceTemplate {
	templates := make([]*ServiceTemplate, 0)

	// PostgreSQL Template
	pgName, _ := NewTemplateName("PostgreSQL")
	pgGitURL, _ := applications.NewGitURL("https://github.com/docker-library/postgres.git", "main", "")
	pgBuildConfig := applications.NewBuildConfig(applications.BuildpackTypeDockerfile)
	pgBuildConfig.SetDockerfileConfig(&applications.DockerfileConfig{
		DockerfilePath: "Dockerfile",
	})

	pgTemplate := NewServiceTemplate(
		pgName,
		"PostgreSQL database server",
		CategoryDatabase,
		"16",
		pgGitURL,
		pgBuildConfig,
		true,
	)
	pgTemplate.SetEnvironmentVariable("POSTGRES_DB", "myapp")
	pgTemplate.SetEnvironmentVariable("POSTGRES_USER", "admin")
	pgTemplate.SetEnvironmentVariable("POSTGRES_PASSWORD", "changeme")
	pgTemplate.AddPort(Port{Port: 5432, Protocol: "tcp", Public: false})
	pgTemplate.AddVolume(Volume{Name: "postgres_data", MountPath: "/var/lib/postgresql/data", Size: "1Gi"})
	templates = append(templates, pgTemplate)

	// Redis Template
	redisName, _ := NewTemplateName("Redis")
	redisGitURL, _ := applications.NewGitURL("https://github.com/docker-library/redis.git", "main", "")
	redisBuildConfig := applications.NewBuildConfig(applications.BuildpackTypeDockerfile)
	redisBuildConfig.SetDockerfileConfig(&applications.DockerfileConfig{
		DockerfilePath: "Dockerfile",
	})

	redisTemplate := NewServiceTemplate(
		redisName,
		"Redis in-memory data store",
		CategoryCache,
		"7",
		redisGitURL,
		redisBuildConfig,
		true,
	)
	redisTemplate.AddPort(Port{Port: 6379, Protocol: "tcp", Public: false})
	redisTemplate.AddVolume(Volume{Name: "redis_data", MountPath: "/data", Size: "500Mi"})
	templates = append(templates, redisTemplate)

	// Supabase Template
	supabaseName, _ := NewTemplateName("Supabase")
	supabaseGitURL, _ := applications.NewGitURL("https://github.com/supabase/supabase.git", "main", "docker")
	supabaseBuildConfig := applications.NewBuildConfig(applications.BuildpackTypeDockerCompose)
	supabaseBuildConfig.SetComposeConfig(&applications.ComposeConfig{
		ComposeFile: "docker-compose.yml",
	})

	supabaseTemplate := NewServiceTemplate(
		supabaseName,
		"Open source Firebase alternative",
		CategoryDatabase,
		"latest",
		supabaseGitURL,
		supabaseBuildConfig,
		true,
	)
	supabaseTemplate.SetEnvironmentVariable("POSTGRES_PASSWORD", "changeme")
	supabaseTemplate.SetEnvironmentVariable("JWT_SECRET", "super-secret-jwt-token-with-at-least-32-characters-long")
	supabaseTemplate.SetEnvironmentVariable("ANON_KEY", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	supabaseTemplate.AddPort(Port{Port: 3000, Protocol: "tcp", Public: true}) // Supabase Studio
	supabaseTemplate.AddPort(Port{Port: 8000, Protocol: "tcp", Public: true}) // Kong API Gateway
	supabaseTemplate.AddVolume(Volume{Name: "supabase_db", MountPath: "/var/lib/postgresql/data", Size: "2Gi"})
	templates = append(templates, supabaseTemplate)

	return templates
}
