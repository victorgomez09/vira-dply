package services

import (
	"context"
	"fmt"

	"github.com/mikrocloud/mikrocloud/internal/config"
	"github.com/mikrocloud/mikrocloud/pkg/containers/build"
	"github.com/mikrocloud/mikrocloud/pkg/containers/manager"
)

type ContainerService struct {
	containerManager manager.ContainerManager
	buildService     *build.BuildService
	config           *config.Config
}

func NewContainerService(cfg *config.Config) (*ContainerService, error) {
	var err error

	// Create container manager
	var containerManager manager.ContainerManager
	switch cfg.Docker.Runtime {
	case "docker":
		containerManager, err = manager.NewDockerManager()
	case "podman":
		containerManager, err = manager.NewPodmanManager()
	default:
		containerManager, err = manager.NewDockerManager() // Default to Docker
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create container manager: %w", err)
	}

	// Create build service
	buildService := build.NewBuildService(containerManager, cfg.Docker.SocketPath)

	return &ContainerService{
		containerManager: containerManager,
		buildService:     buildService,
		config:           cfg,
	}, nil
}

// Container lifecycle operations
func (cs *ContainerService) StartContainer(ctx context.Context, containerID string) error {
	return cs.containerManager.Start(ctx, containerID)
}

func (cs *ContainerService) StopContainer(ctx context.Context, containerID string) error {
	return cs.containerManager.Stop(ctx, containerID)
}

func (cs *ContainerService) RestartContainer(ctx context.Context, containerID string) error {
	return cs.containerManager.Restart(ctx, containerID)
}

func (cs *ContainerService) DeleteContainer(ctx context.Context, containerID string) error {
	return cs.containerManager.Delete(ctx, containerID)
}

// Container management
func (cs *ContainerService) CreateContainer(ctx context.Context, config manager.ContainerConfig) (string, error) {
	return cs.containerManager.Create(ctx, config)
}

func (cs *ContainerService) ListContainers(ctx context.Context) ([]manager.ContainerInfo, error) {
	return cs.containerManager.List(ctx)
}

func (cs *ContainerService) InspectContainer(ctx context.Context, containerID string) (*manager.ContainerInfo, error) {
	return cs.containerManager.Inspect(ctx, containerID)
}

// Logging
func (cs *ContainerService) StreamContainerLogs(ctx context.Context, containerID string, follow bool) error {
	logStream, err := cs.containerManager.StreamLogs(ctx, containerID, follow)
	if err != nil {
		return err
	}
	defer logStream.Close()

	// TODO: Process and forward logs to appropriate channels (WebSocket, etc.)
	return nil
}

// Image operations
func (cs *ContainerService) PullImage(ctx context.Context, image string) error {
	return cs.containerManager.PullImage(ctx, image)
}

func (cs *ContainerService) BuildImage(ctx context.Context, buildRequest build.BuildRequest) (*build.BuildResult, error) {
	return cs.buildService.BuildImage(ctx, buildRequest)
}

// Helper methods
func (cs *ContainerService) GetRuntimeInfo() map[string]interface{} {
	return map[string]interface{}{
		"runtime":     cs.config.Docker.Runtime,
		"socket_path": cs.config.Docker.SocketPath,
		"rootless":    cs.config.Docker.Rootless,
		"build_dir":   cs.config.Docker.BuildDir,
	}
}
