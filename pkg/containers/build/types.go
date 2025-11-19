package build

import (
	"fmt"
)

type BuildpackType string

const (
	Nixpacks       BuildpackType = "nixpacks"
	Static         BuildpackType = "static"
	DockerfileType BuildpackType = "dockerfile"
	DockerCompose  BuildpackType = "docker-compose"
)

type BuildRequest struct {
	ID            string
	GitRepo       string
	GitBranch     string
	ContextRoot   string
	BuildpackType BuildpackType
	Environment   map[string]string
	ImageTag      string

	// Buildpack-specific configurations
	NixpacksConfig   *NixpacksConfig   `json:"nixpacks_config,omitempty"`
	StaticConfig     *StaticConfig     `json:"static_config,omitempty"`
	DockerfileConfig *DockerfileConfig `json:"dockerfile_config,omitempty"`
	ComposeConfig    *ComposeConfig    `json:"compose_config,omitempty"`

	// Optional callback for streaming logs in real-time
	LogCallback func(log string) `json:"-"`
}

type BuildResult struct {
	Success   bool
	ImageTag  string
	BuildLogs string
	Error     string
}

type BuildpackConfig interface {
	GetDockerfile() string
	Validate() error
}

type NixpacksConfig struct {
	StartCommand string            `json:"start_command,omitempty"`
	BuildCommand string            `json:"build_command,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
}

func (n *NixpacksConfig) GetDockerfile() string {
	return fmt.Sprintf(`
FROM nixpacks/nixpacks:latest as builder
WORKDIR /app
COPY . .
RUN nixpacks build . --name app

FROM nixpacks/nixpacks:runtime
COPY --from=builder /app /app
WORKDIR /app
%s
`, n.getStartCommand())
}

func (n *NixpacksConfig) getStartCommand() string {
	if n.StartCommand != "" {
		return fmt.Sprintf("CMD [%s]", n.StartCommand)
	}
	return "CMD nixpacks start"
}

func (n *NixpacksConfig) Validate() error {
	return nil // Nixpacks handles most validation
}

type StaticConfig struct {
	BuildCommand string `json:"build_command,omitempty"`
	OutputDir    string `json:"output_dir,omitempty"`
	NginxConfig  string `json:"nginx_config,omitempty"`
}

func (s *StaticConfig) GetDockerfile() string {
	outputDir := s.OutputDir
	if outputDir == "" {
		outputDir = "dist"
	}

	buildCmd := s.BuildCommand
	if buildCmd == "" {
		buildCmd = "npm run build"
	}

	return fmt.Sprintf(`
FROM node:18-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN %s

FROM nginx:alpine
COPY --from=builder /app/%s /usr/share/nginx/html
%s
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
`, buildCmd, outputDir, s.getNginxConfig())
}

func (s *StaticConfig) getNginxConfig() string {
	if s.NginxConfig != "" {
		return fmt.Sprintf("COPY %s /etc/nginx/conf.d/default.conf", s.NginxConfig)
	}
	return ""
}

func (s *StaticConfig) Validate() error {
	if s.OutputDir == "" {
		s.OutputDir = "dist"
	}
	return nil
}

type DockerfileConfig struct {
	DockerfilePath string            `json:"dockerfile_path,omitempty"`
	BuildArgs      map[string]string `json:"build_args,omitempty"`
	Target         string            `json:"target,omitempty"`
}

func (d *DockerfileConfig) GetDockerfile() string {
	// For Dockerfile builds, we use the existing Dockerfile
	return ""
}

func (d *DockerfileConfig) Validate() error {
	if d.DockerfilePath == "" {
		d.DockerfilePath = "Dockerfile"
	}
	return nil
}

type ComposeConfig struct {
	ComposeFile string `json:"compose_file,omitempty"`
	Service     string `json:"service,omitempty"`
}

func (c *ComposeConfig) GetDockerfile() string {
	// For compose builds, we use the compose file
	return ""
}

func (c *ComposeConfig) Validate() error {
	if c.ComposeFile == "" {
		c.ComposeFile = "docker-compose.yml"
	}
	return nil
}
