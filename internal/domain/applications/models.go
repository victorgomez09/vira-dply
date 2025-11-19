package applications

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DeploymentSource struct {
	Type     DeploymentSourceType `json:"type"`
	GitRepo  *GitRepoSource       `json:"git_repo,omitempty"`
	Registry *RegistrySource      `json:"registry,omitempty"`
	Upload   *UploadSource        `json:"upload,omitempty"`
}

type DeploymentSourceType string

const (
	DeploymentSourceTypeGit      DeploymentSourceType = "git"
	DeploymentSourceTypeRegistry DeploymentSourceType = "registry"
	DeploymentSourceTypeUpload   DeploymentSourceType = "upload"
)

// GitRepoSource represents a Git repository source with enhanced validation
type GitRepoSource struct {
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	Path     string `json:"path,omitempty"`      // Context root within repo (subdir)
	BasePath string `json:"base_path,omitempty"` // Base path for Dockerfile/Compose file
	Token    string `json:"token,omitempty"`     // For private repos
}

// GitURL is a value object for Git repository URLs with enhanced validation
type GitURL struct {
	value       string
	branch      string
	contextRoot string
}

func NewGitURL(url, branch, contextRoot string) (*GitURL, error) {
	if url == "" {
		return nil, fmt.Errorf("git URL cannot be empty")
	}

	if branch == "" {
		branch = "main" // Default branch
	}

	return &GitURL{
		value:       url,
		branch:      branch,
		contextRoot: contextRoot,
	}, nil
}

func (g *GitURL) URL() string {
	return g.value
}

func (g *GitURL) Branch() string {
	return g.branch
}

func (g *GitURL) ContextRoot() string {
	return g.contextRoot
}

func (g *GitURL) ToGitRepoSource() *GitRepoSource {
	return &GitRepoSource{
		URL:    g.value,
		Branch: g.branch,
		Path:   g.contextRoot,
	}
}

type RegistrySource struct {
	Image string `json:"image"`
	Tag   string `json:"tag"`
}

type UploadSource struct {
	Filename string `json:"filename"`
	FilePath string `json:"file_path"`
}

// BuildConfig represents enhanced build configuration with specific buildpack configs
type BuildConfig struct {
	buildpackType BuildpackType
	nixpacks      *NixpacksConfig
	static        *StaticConfig
	dockerfile    *DockerfileConfig
	compose       *ComposeConfig
}

type NixpacksConfig struct {
	StartCommand string            `json:"start_command,omitempty"`
	BuildCommand string            `json:"build_command,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
}

type StaticConfig struct {
	BuildCommand string `json:"build_command,omitempty"`
	OutputDir    string `json:"output_dir,omitempty"`
	NginxConfig  string `json:"nginx_config,omitempty"`
}

type DockerfileConfig struct {
	DockerfilePath string            `json:"dockerfile_path,omitempty"` // Path to Dockerfile (e.g., "Dockerfile")
	BasePath       string            `json:"base_path,omitempty"`       // Base directory containing Dockerfile (e.g., "./backend")
	BuildArgs      map[string]string `json:"build_args,omitempty"`
	Target         string            `json:"target,omitempty"`
}

type ComposeConfig struct {
	ComposeFile string `json:"compose_file,omitempty"` // Path to compose file (e.g., "docker-compose.yml")
	BasePath    string `json:"base_path,omitempty"`    // Base directory containing compose file
	Service     string `json:"service,omitempty"`
}

func NewBuildConfig(buildpackType BuildpackType) *BuildConfig {
	return &BuildConfig{
		buildpackType: buildpackType,
	}
}

func (bc *BuildConfig) SetNixpacksConfig(config *NixpacksConfig) {
	bc.buildpackType = BuildpackTypeNixpacks
	bc.nixpacks = config
}

func (bc *BuildConfig) SetStaticConfig(config *StaticConfig) {
	bc.buildpackType = BuildpackTypeStatic
	bc.static = config
}

func (bc *BuildConfig) SetDockerfileConfig(config *DockerfileConfig) {
	bc.buildpackType = BuildpackTypeDockerfile
	bc.dockerfile = config
}

func (bc *BuildConfig) SetComposeConfig(config *ComposeConfig) {
	bc.buildpackType = BuildpackTypeDockerCompose
	bc.compose = config
}

func (bc *BuildConfig) BuildpackType() BuildpackType {
	return bc.buildpackType
}

func (bc *BuildConfig) NixpacksConfig() *NixpacksConfig {
	return bc.nixpacks
}

func (bc *BuildConfig) StaticConfig() *StaticConfig {
	return bc.static
}

func (bc *BuildConfig) DockerfileConfig() *DockerfileConfig {
	return bc.dockerfile
}

func (bc *BuildConfig) ComposeConfig() *ComposeConfig {
	return bc.compose
}

// Legacy BuildpackConfig for backward compatibility
type BuildpackConfig struct {
	Type   BuildpackType `json:"type"`
	Config any           `json:"config,omitempty"`
}

type PortMapping struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port,omitempty"`
	Protocol      string `json:"protocol"` // tcp, udp
}

type Application struct {
	id               ApplicationID
	name             ApplicationName
	description      string
	projectID        uuid.UUID
	environmentID    uuid.UUID
	deploymentSource DeploymentSource
	domain           string
	generatedDomain  string
	exposedPorts     []int
	portMappings     []PortMapping
	buildpack        *BuildConfig
	envVars          map[string]string
	autoDeploy       bool
	status           ApplicationStatus
	createdAt        time.Time
	updatedAt        time.Time
}

type ApplicationID struct {
	value string
}

func NewApplicationID() ApplicationID {
	return ApplicationID{value: uuid.New().String()}
}

func ApplicationIDFromString(s string) (ApplicationID, error) {
	if s == "" {
		return ApplicationID{}, fmt.Errorf("application ID cannot be empty")
	}
	return ApplicationID{value: s}, nil
}

func (id ApplicationID) String() string {
	return id.value
}

type ApplicationName struct {
	value string
}

func NewApplicationName(name string) (ApplicationName, error) {
	if name == "" {
		return ApplicationName{}, fmt.Errorf("application name cannot be empty")
	}
	if len(name) > 100 {
		return ApplicationName{}, fmt.Errorf("application name cannot exceed 100 characters")
	}
	return ApplicationName{value: name}, nil
}

func (n ApplicationName) String() string {
	return n.value
}

type BuildpackType string

const (
	BuildpackTypeNixpacks      BuildpackType = "nixpacks"
	BuildpackTypeStatic        BuildpackType = "static"
	BuildpackTypeDockerfile    BuildpackType = "dockerfile"
	BuildpackTypeDockerCompose BuildpackType = "docker-compose"
	BuildpackTypeBuildpacks    BuildpackType = "buildpacks"
)

type ApplicationStatus string

const (
	ApplicationStatusCreated   ApplicationStatus = "created"
	ApplicationStatusBuilding  ApplicationStatus = "building"
	ApplicationStatusDeploying ApplicationStatus = "deploying"
	ApplicationStatusRunning   ApplicationStatus = "running"
	ApplicationStatusStopped   ApplicationStatus = "stopped"
	ApplicationStatusFailed    ApplicationStatus = "failed"
)

func NewApplication(
	name ApplicationName,
	description string,
	projectID, environmentID uuid.UUID,
	deploymentSource DeploymentSource,
	buildpack *BuildConfig,
) *Application {
	now := time.Now()
	return &Application{
		id:               NewApplicationID(),
		name:             name,
		description:      description,
		projectID:        projectID,
		environmentID:    environmentID,
		deploymentSource: deploymentSource,
		buildpack:        buildpack,
		envVars:          make(map[string]string),
		exposedPorts:     []int{},
		portMappings:     []PortMapping{},
		autoDeploy:       true,
		status:           ApplicationStatusCreated,
		createdAt:        now,
		updatedAt:        now,
	}
}

func (a *Application) ID() ApplicationID {
	return a.id
}

func (a *Application) Name() ApplicationName {
	return a.name
}

func (a *Application) Description() string {
	return a.description
}

func (a *Application) ProjectID() uuid.UUID {
	return a.projectID
}

func (a *Application) EnvironmentID() uuid.UUID {
	return a.environmentID
}

func (a *Application) DeploymentSource() DeploymentSource {
	return a.deploymentSource
}

func (a *Application) RepoURL() string {
	if a.deploymentSource.Type == DeploymentSourceTypeGit && a.deploymentSource.GitRepo != nil {
		return a.deploymentSource.GitRepo.URL
	}
	return ""
}

func (a *Application) RepoBranch() string {
	if a.deploymentSource.Type == DeploymentSourceTypeGit && a.deploymentSource.GitRepo != nil {
		return a.deploymentSource.GitRepo.Branch
	}
	return ""
}

func (a *Application) RepoPath() string {
	if a.deploymentSource.Type == DeploymentSourceTypeGit && a.deploymentSource.GitRepo != nil {
		return a.deploymentSource.GitRepo.Path
	}
	return ""
}

func (a *Application) BasePath() string {
	if a.deploymentSource.Type == DeploymentSourceTypeGit && a.deploymentSource.GitRepo != nil {
		return a.deploymentSource.GitRepo.BasePath
	}
	return ""
}

func (a *Application) Domain() string {
	return a.domain
}

func (a *Application) GeneratedDomain() string {
	return a.generatedDomain
}

func (a *Application) ExposedPorts() []int {
	return a.exposedPorts
}

func (a *Application) PortMappings() []PortMapping {
	return a.portMappings
}

func (a *Application) BuildpackType() BuildpackType {
	if a.buildpack != nil {
		return a.buildpack.BuildpackType()
	}
	return BuildpackTypeNixpacks // default
}

func (a *Application) Config() string {
	// For backward compatibility, serialize the entire config
	if a.buildpack == nil {
		return "{}"
	}

	switch a.buildpack.BuildpackType() {
	case BuildpackTypeNixpacks:
		if config := a.buildpack.NixpacksConfig(); config != nil {
			if configJSON, err := json.Marshal(config); err == nil {
				return string(configJSON)
			}
		}
	case BuildpackTypeStatic:
		if config := a.buildpack.StaticConfig(); config != nil {
			if configJSON, err := json.Marshal(config); err == nil {
				return string(configJSON)
			}
		}
	case BuildpackTypeDockerfile:
		if config := a.buildpack.DockerfileConfig(); config != nil {
			if configJSON, err := json.Marshal(config); err == nil {
				return string(configJSON)
			}
		}
	case BuildpackTypeDockerCompose:
		if config := a.buildpack.ComposeConfig(); config != nil {
			if configJSON, err := json.Marshal(config); err == nil {
				return string(configJSON)
			}
		}
	}
	return "{}"
}

func (a *Application) Buildpack() *BuildConfig {
	return a.buildpack
}

func (a *Application) EnvVars() map[string]string {
	result := make(map[string]string)
	for k, v := range a.envVars {
		result[k] = v
	}
	return result
}

func (a *Application) AutoDeploy() bool {
	return a.autoDeploy
}

func (a *Application) Status() ApplicationStatus {
	return a.status
}

func (a *Application) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Application) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a *Application) UpdateDescription(description string) {
	a.description = description
	a.updatedAt = time.Now()
}

func (a *Application) SetDeploymentSource(source DeploymentSource) {
	a.deploymentSource = source
	a.updatedAt = time.Now()
}

func (a *Application) SetRepoURL(repoURL string) {
	if a.deploymentSource.Type == DeploymentSourceTypeGit {
		if a.deploymentSource.GitRepo == nil {
			a.deploymentSource.GitRepo = &GitRepoSource{}
		}
		a.deploymentSource.GitRepo.URL = repoURL
		a.updatedAt = time.Now()
	}
}

func (a *Application) SetRepoBranch(branch string) {
	if a.deploymentSource.Type == DeploymentSourceTypeGit {
		if a.deploymentSource.GitRepo == nil {
			a.deploymentSource.GitRepo = &GitRepoSource{}
		}
		a.deploymentSource.GitRepo.Branch = branch
		a.updatedAt = time.Now()
	}
}

func (a *Application) SetRepoPath(path string) {
	if a.deploymentSource.Type == DeploymentSourceTypeGit {
		if a.deploymentSource.GitRepo == nil {
			a.deploymentSource.GitRepo = &GitRepoSource{}
		}
		a.deploymentSource.GitRepo.Path = path
		a.updatedAt = time.Now()
	}
}

func (a *Application) SetDomain(domain string) {
	a.domain = domain
	a.updatedAt = time.Now()
}

func (a *Application) SetGeneratedDomain(domain string) {
	a.generatedDomain = domain
	a.updatedAt = time.Now()
}

func (a *Application) UpdateName(name ApplicationName) {
	a.name = name
	a.updatedAt = time.Now()
}

func (a *Application) SetExposedPorts(ports []int) error {
	for _, port := range ports {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid port: %d", port)
		}
	}
	a.exposedPorts = ports
	a.updatedAt = time.Now()
	return nil
}

func (a *Application) AddPortMapping(containerPort, hostPort int, protocol string) error {
	if containerPort < 1 || containerPort > 65535 {
		return fmt.Errorf("invalid container port: %d", containerPort)
	}
	if hostPort != 0 && (hostPort < 1 || hostPort > 65535) {
		return fmt.Errorf("invalid host port: %d", hostPort)
	}
	if protocol != "tcp" && protocol != "udp" {
		protocol = "tcp"
	}

	a.portMappings = append(a.portMappings, PortMapping{
		ContainerPort: containerPort,
		HostPort:      hostPort,
		Protocol:      protocol,
	})
	a.updatedAt = time.Now()
	return nil
}

func (a *Application) SetPortMappings(mappings []PortMapping) error {
	for _, mapping := range mappings {
		if mapping.ContainerPort < 1 || mapping.ContainerPort > 65535 {
			return fmt.Errorf("invalid container port: %d", mapping.ContainerPort)
		}
		if mapping.HostPort != 0 && (mapping.HostPort < 1 || mapping.HostPort > 65535) {
			return fmt.Errorf("invalid host port: %d", mapping.HostPort)
		}
		if mapping.Protocol != "tcp" && mapping.Protocol != "udp" {
			mapping.Protocol = "tcp"
		}
	}
	a.portMappings = mappings
	a.updatedAt = time.Now()
	return nil
}

func (a *Application) SetBuildpackType(buildpackType BuildpackType) {
	if a.buildpack == nil {
		a.buildpack = NewBuildConfig(buildpackType)
	} else {
		// Create new config with the new type, preserving existing config if possible
		newConfig := NewBuildConfig(buildpackType)
		// TODO: Add logic to preserve config when switching types if needed
		a.buildpack = newConfig
	}
	a.updatedAt = time.Now()
}

func (a *Application) UpdateConfig(config string) {
	if a.buildpack == nil {
		return
	}

	var configData interface{}
	if err := json.Unmarshal([]byte(config), &configData); err != nil {
		return
	}

	// Update specific config based on buildpack type
	switch a.buildpack.BuildpackType() {
	case BuildpackTypeNixpacks:
		if nixpacksConfig, ok := configData.(*NixpacksConfig); ok {
			a.buildpack.SetNixpacksConfig(nixpacksConfig)
		}
	case BuildpackTypeStatic:
		if staticConfig, ok := configData.(*StaticConfig); ok {
			a.buildpack.SetStaticConfig(staticConfig)
		}
	case BuildpackTypeDockerfile:
		if dockerfileConfig, ok := configData.(*DockerfileConfig); ok {
			a.buildpack.SetDockerfileConfig(dockerfileConfig)
		}
	case BuildpackTypeDockerCompose:
		if composeConfig, ok := configData.(*ComposeConfig); ok {
			a.buildpack.SetComposeConfig(composeConfig)
		}
	}
	a.updatedAt = time.Now()
}

func (a *Application) SetBuildpack(buildpack *BuildConfig) {
	a.buildpack = buildpack
	a.updatedAt = time.Now()
}

func (a *Application) SetEnvVar(key, value string) {
	if a.envVars == nil {
		a.envVars = make(map[string]string)
	}
	a.envVars[key] = value
	a.updatedAt = time.Now()
}

func (a *Application) RemoveEnvVar(key string) {
	if a.envVars != nil {
		delete(a.envVars, key)
		a.updatedAt = time.Now()
	}
}

func (a *Application) SetEnvVars(envVars map[string]string) {
	a.envVars = make(map[string]string)
	for k, v := range envVars {
		a.envVars[k] = v
	}
	a.updatedAt = time.Now()
}

func (a *Application) SetAutoDeploy(autoDeploy bool) {
	a.autoDeploy = autoDeploy
	a.updatedAt = time.Now()
}

func (a *Application) ChangeStatus(status ApplicationStatus) {
	a.status = status
	a.updatedAt = time.Now()
}

func (a *Application) CanDeploy() error {
	switch a.status {
	case ApplicationStatusBuilding:
		return fmt.Errorf("application is currently building")
	case ApplicationStatusDeploying:
		return fmt.Errorf("application is currently deploying")
	default:
		return nil
	}
}

func (a *Application) CanStop() error {
	if a.status != ApplicationStatusRunning {
		return fmt.Errorf("application is not running")
	}
	return nil
}

func ReconstructApplication(
	id ApplicationID,
	name ApplicationName,
	description string,
	projectID, environmentID uuid.UUID,
	deploymentSource DeploymentSource,
	domain string,
	generatedDomain string,
	exposedPorts []int,
	portMappings []PortMapping,
	buildpack *BuildConfig,
	envVars map[string]string,
	autoDeploy bool,
	status ApplicationStatus,
	createdAt, updatedAt time.Time,
) *Application {
	if envVars == nil {
		envVars = make(map[string]string)
	}
	if exposedPorts == nil {
		exposedPorts = []int{}
	}
	if portMappings == nil {
		portMappings = []PortMapping{}
	}
	return &Application{
		id:               id,
		name:             name,
		description:      description,
		projectID:        projectID,
		environmentID:    environmentID,
		deploymentSource: deploymentSource,
		domain:           domain,
		generatedDomain:  generatedDomain,
		exposedPorts:     exposedPorts,
		portMappings:     portMappings,
		buildpack:        buildpack,
		envVars:          envVars,
		autoDeploy:       autoDeploy,
		status:           status,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

// Helper functions for creating deployment sources
func NewGitDeploymentSource(url, branch, path, basePath string) DeploymentSource {
	if branch == "" {
		branch = "main"
	}
	return DeploymentSource{
		Type: DeploymentSourceTypeGit,
		GitRepo: &GitRepoSource{
			URL:      url,
			Branch:   branch,
			Path:     path,
			BasePath: basePath,
		},
	}
}

func NewRegistryDeploymentSource(image, tag string) DeploymentSource {
	if tag == "" {
		tag = "latest"
	}
	return DeploymentSource{
		Type: DeploymentSourceTypeRegistry,
		Registry: &RegistrySource{
			Image: image,
			Tag:   tag,
		},
	}
}

func NewUploadDeploymentSource(filename, filePath string) DeploymentSource {
	return DeploymentSource{
		Type: DeploymentSourceTypeUpload,
		Upload: &UploadSource{
			Filename: filename,
			FilePath: filePath,
		},
	}
}

// Helper functions for creating buildpack configs
func NewLegacyBuildpackConfig(buildpackType BuildpackType, config interface{}) *BuildConfig {
	buildConfig := NewBuildConfig(buildpackType)

	if config == nil {
		return buildConfig
	}

	// Handle JSON map conversion from HTTP requests
	configMap, isMap := config.(map[string]interface{})

	// Try to convert legacy config to new format
	switch buildpackType {
	case BuildpackTypeNixpacks:
		if nixConfig, ok := config.(*NixpacksConfig); ok {
			buildConfig.SetNixpacksConfig(nixConfig)
		} else if isMap {
			nixConfig := &NixpacksConfig{}
			if startCmd, ok := configMap["start_command"].(string); ok {
				nixConfig.StartCommand = startCmd
			}
			if buildCmd, ok := configMap["build_command"].(string); ok {
				nixConfig.BuildCommand = buildCmd
			}
			if vars, ok := configMap["variables"].(map[string]interface{}); ok {
				nixConfig.Variables = make(map[string]string)
				for k, v := range vars {
					if strVal, ok := v.(string); ok {
						nixConfig.Variables[k] = strVal
					}
				}
			}
			buildConfig.SetNixpacksConfig(nixConfig)
		}
	case BuildpackTypeStatic:
		if staticConfig, ok := config.(*StaticConfig); ok {
			buildConfig.SetStaticConfig(staticConfig)
		} else if isMap {
			staticConfig := &StaticConfig{}
			if buildCmd, ok := configMap["build_command"].(string); ok {
				staticConfig.BuildCommand = buildCmd
			}
			if outputDir, ok := configMap["output_dir"].(string); ok {
				staticConfig.OutputDir = outputDir
			}
			if nginxCfg, ok := configMap["nginx_config"].(string); ok {
				staticConfig.NginxConfig = nginxCfg
			}
			buildConfig.SetStaticConfig(staticConfig)
		}
	case BuildpackTypeDockerfile:
		if dockerConfig, ok := config.(*DockerfileConfig); ok {
			buildConfig.SetDockerfileConfig(dockerConfig)
		} else if isMap {
			dockerConfig := &DockerfileConfig{}
			if dockerfilePath, ok := configMap["dockerfile_path"].(string); ok {
				dockerConfig.DockerfilePath = dockerfilePath
			}
			if basePath, ok := configMap["base_path"].(string); ok {
				dockerConfig.BasePath = basePath
			}
			if target, ok := configMap["target"].(string); ok {
				dockerConfig.Target = target
			}
			if buildArgs, ok := configMap["build_args"].(map[string]interface{}); ok {
				dockerConfig.BuildArgs = make(map[string]string)
				for k, v := range buildArgs {
					if strVal, ok := v.(string); ok {
						dockerConfig.BuildArgs[k] = strVal
					}
				}
			}
			buildConfig.SetDockerfileConfig(dockerConfig)
		}
	case BuildpackTypeDockerCompose:
		if composeConfig, ok := config.(*ComposeConfig); ok {
			buildConfig.SetComposeConfig(composeConfig)
		} else if isMap {
			composeConfig := &ComposeConfig{}
			if composeFile, ok := configMap["compose_file"].(string); ok {
				composeConfig.ComposeFile = composeFile
			}
			if basePath, ok := configMap["base_path"].(string); ok {
				composeConfig.BasePath = basePath
			}
			if service, ok := configMap["service"].(string); ok {
				composeConfig.Service = service
			}
			buildConfig.SetComposeConfig(composeConfig)
		}
	}

	return buildConfig
}
