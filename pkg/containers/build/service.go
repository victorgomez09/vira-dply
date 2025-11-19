package build

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mikrocloud/mikrocloud/pkg/containers/manager"
)

type BuildService struct {
	containerManager      manager.ContainerManager
	containerEngineSocket string // Path to Docker/Podman socket
}

func NewBuildService(containerManager manager.ContainerManager, dockerSocket string) *BuildService {
	if dockerSocket == "" {
		dockerSocket = "/var/run/docker.sock" // Default Docker socket
	}
	return &BuildService{
		containerManager:      containerManager,
		containerEngineSocket: dockerSocket,
	}
}

func (bs *BuildService) BuildImage(ctx context.Context, request BuildRequest) (*BuildResult, error) {
	// Create a unique build container name
	buildContainerName := fmt.Sprintf("mikrocloud-build-%s", request.ID)

	// All building happens inside this helper container
	// The helper container has access to the host Docker daemon via socket mount
	switch request.BuildpackType {
	case Nixpacks:
		return bs.buildWithNixpacks(ctx, request, buildContainerName)
	case Static:
		return bs.buildStatic(ctx, request, buildContainerName)
	case DockerfileType:
		return bs.buildWithDockerfile(ctx, request, buildContainerName)
	case DockerCompose:
		return bs.buildWithCompose(ctx, request, buildContainerName)
	default:
		return &BuildResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported buildpack type: %s", request.BuildpackType),
		}, nil
	}
}

// Helper function to create a build helper container that clones repo and executes build commands
func (bs *BuildService) createBuildHelper(ctx context.Context, image, containerName string, commands []string, request BuildRequest) (*BuildResult, error) {
	// Environment variables for the build helper
	env := map[string]string{
		"GIT_REPO":     request.GitRepo,
		"GIT_BRANCH":   request.GitBranch,
		"CONTEXT_ROOT": request.ContextRoot,
		"IMAGE_TAG":    request.ImageTag,
		"BUILD_ID":     request.ID,
	}

	// Add user-defined environment variables
	for k, v := range request.Environment {
		env[k] = v
	}

	// Prepare the command to run inside the helper container
	// This will clone the repo and execute the build commands
	fullCommand := []string{
		"/bin/sh", "-c",
		bs.generateBuildScript(commands, request),
	}

	// Pull the build helper image if not present
	if err := bs.containerManager.PullImage(ctx, image); err != nil {
		return &BuildResult{Success: false, Error: fmt.Sprintf("failed to pull build helper image: %v", err)}, nil
	}

	containerConfig := manager.ContainerConfig{
		Image:       image,
		Name:        containerName,
		Command:     fullCommand,
		Environment: env,
		Volumes: map[string]string{
			bs.containerEngineSocket: "/var/run/docker.sock", // Mount Docker socket for building images
		},
		WorkingDir: "/workspace",
		AutoRemove: true, // Automatically remove container when it exits
	}

	containerID, err := bs.containerManager.Create(ctx, containerConfig)
	if err != nil {
		return &BuildResult{Success: false, Error: fmt.Sprintf("failed to create build helper: %v", err)}, nil
	}

	// Start the build process
	if err := bs.containerManager.Start(ctx, containerID); err != nil {
		return &BuildResult{Success: false, Error: fmt.Sprintf("failed to start build: %v", err)}, nil
	}

	// Stream and capture build logs
	logStream, err := bs.containerManager.StreamLogs(ctx, containerID, true)
	if err != nil {
		return &BuildResult{Success: false, Error: fmt.Sprintf("failed to get build logs: %v", err)}, nil
	}
	defer logStream.Close()

	// Stream logs in real-time if callback is provided
	var allLogs strings.Builder
	done := make(chan error, 1)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := logStream.Read(buf)
			if n > 0 {
				logChunk := string(buf[:n])
				allLogs.WriteString(logChunk)

				// Call the callback for real-time streaming
				if request.LogCallback != nil {
					request.LogCallback(logChunk)
				}
			}
			if err != nil {
				if err != io.EOF {
					done <- err
				} else {
					done <- nil
				}
				return
			}
		}
	}()

	// Wait for container to finish and get exit code
	exitCode, err := bs.containerManager.Wait(ctx, containerID)

	// Wait for log streaming to complete
	logErr := <-done

	if err != nil {
		return &BuildResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to wait for container: %v", err),
			BuildLogs: allLogs.String(),
		}, nil
	}

	if logErr != nil {
		return &BuildResult{
			Success:   exitCode == 0,
			Error:     fmt.Sprintf("failed to read build logs: %v", logErr),
			BuildLogs: allLogs.String(),
		}, nil
	}

	// Determine success based on exit code
	success := exitCode == 0
	errorMsg := ""
	if !success {
		errorMsg = fmt.Sprintf("build failed with exit code %d", exitCode)
	}

	return &BuildResult{
		Success:   success,
		ImageTag:  request.ImageTag,
		BuildLogs: allLogs.String(),
		Error:     errorMsg,
	}, nil
}

// Generate a shell script that will run inside the build helper container
func (bs *BuildService) generateBuildScript(commands []string, request BuildRequest) string {
	script := []string{
		"set -e", // Exit on any error
		"echo '=== Starting build process ==='",
		"",
		"# Clone the repository",
		"echo '=== Cloning repository ==='",
	}

	// Add git clone command
	if request.GitBranch != "" {
		cloneCmd := fmt.Sprintf("git clone -b %s %s /workspace/source", request.GitBranch, request.GitRepo)
		script = append(script, fmt.Sprintf("echo 'Running: %s'", cloneCmd))
		script = append(script, cloneCmd)
	} else {
		cloneCmd := fmt.Sprintf("git clone %s /workspace/source", request.GitRepo)
		script = append(script, fmt.Sprintf("echo 'Running: %s'", cloneCmd))
		script = append(script, cloneCmd)
	}

	// Change to context directory if specified
	if request.ContextRoot != "" {
		cdCmd := fmt.Sprintf("cd /workspace/source/%s", request.ContextRoot)
		script = append(script, fmt.Sprintf("echo 'Running: %s'", cdCmd))
		script = append(script, cdCmd)
	} else {
		script = append(script, "echo 'Running: cd /workspace/source'")
		script = append(script, "cd /workspace/source")
	}

	script = append(script, "")
	script = append(script, "echo '=== Executing build commands ==='")

	// Add build commands with logging
	for _, cmd := range commands {
		script = append(script, fmt.Sprintf("echo 'Running: %s'", cmd))
		script = append(script, cmd)
	}

	return strings.Join(script, "\n")
}

func (bs *BuildService) buildWithNixpacks(ctx context.Context, request BuildRequest, containerName string) (*BuildResult, error) {
	config := request.NixpacksConfig
	if config == nil {
		config = &NixpacksConfig{}
	}

	if err := config.Validate(); err != nil {
		return &BuildResult{Success: false, Error: err.Error()}, nil
	}

	// Commands to run inside the nixpacks build helper
	commands := []string{
		"echo 'Building with Nixpacks...'",
		fmt.Sprintf("nixpacks build . --name %s", request.ImageTag),
	}

	// Use nixpacks image as the build helper
	return bs.createBuildHelper(ctx, "railwayapp/nixpacks:latest", containerName, commands, request)
}

func (bs *BuildService) buildStatic(ctx context.Context, request BuildRequest, containerName string) (*BuildResult, error) {
	config := request.StaticConfig
	if config == nil {
		config = &StaticConfig{}
	}

	if err := config.Validate(); err != nil {
		return &BuildResult{Success: false, Error: err.Error()}, nil
	}

	// Generate Dockerfile content inside the container
	dockerfileContent := strings.ReplaceAll(config.GetDockerfile(), "\n", "\\n")
	dockerfileContent = strings.ReplaceAll(dockerfileContent, "\"", "\\\"")

	commands := []string{
		"echo 'Building static site...'",
		fmt.Sprintf("printf \"%s\" > Dockerfile", dockerfileContent),
		fmt.Sprintf("docker build -t %s .", request.ImageTag),
	}

	// Use mikrocloud-builder with git and Docker CLI
	return bs.createBuildHelper(ctx, "ghcr.io/fnprog/mikrocloud/mikrocloud-builder:latest", containerName, commands, request)
}

func (bs *BuildService) buildWithDockerfile(ctx context.Context, request BuildRequest, containerName string) (*BuildResult, error) {
	config := request.DockerfileConfig
	if config == nil {
		config = &DockerfileConfig{}
	}

	if err := config.Validate(); err != nil {
		return &BuildResult{Success: false, Error: err.Error()}, nil
	}

	// Build using existing Dockerfile
	dockerfilePath := config.DockerfilePath
	if dockerfilePath == "" {
		dockerfilePath = "Dockerfile"
	}

	buildArgs := ""
	for key, value := range config.BuildArgs {
		buildArgs += fmt.Sprintf(" --build-arg %s=%s", key, value)
	}

	targetFlag := ""
	if config.Target != "" {
		targetFlag = fmt.Sprintf(" --target %s", config.Target)
	}

	commands := []string{
		"echo 'Building with Dockerfile...'",
		fmt.Sprintf("docker build -f %s%s%s -t %s .", dockerfilePath, buildArgs, targetFlag, request.ImageTag),
	}

	// Use mikrocloud-builder with Docker CLI and git
	return bs.createBuildHelper(ctx, "ghcr.io/fnprog/mikrocloud/mikrocloud-builder:latest", containerName, commands, request)
}

func (bs *BuildService) buildWithCompose(ctx context.Context, request BuildRequest, containerName string) (*BuildResult, error) {
	config := request.ComposeConfig
	if config == nil {
		config = &ComposeConfig{}
	}

	if err := config.Validate(); err != nil {
		return &BuildResult{Success: false, Error: err.Error()}, nil
	}

	composeFile := config.ComposeFile
	if composeFile == "" {
		composeFile = "docker-compose.yml"
	}

	var buildCmd string
	if config.Service != "" {
		buildCmd = fmt.Sprintf("docker-compose -f %s build %s", composeFile, config.Service)
	} else {
		buildCmd = fmt.Sprintf("docker-compose -f %s build", composeFile)
	}

	commands := []string{
		"echo 'Building with Docker Compose...'",
		buildCmd,
	}

	// Use mikrocloud-builder with Docker Compose
	return bs.createBuildHelper(ctx, "ghcr.io/fnprog/mikrocloud/mikrocloud-builder:latest", containerName, commands, request)
}
