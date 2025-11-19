package manager

import (
	"context"
	"fmt"
	"io"
)

type ContainerManager interface {
	// Container lifecycle operations
	Start(ctx context.Context, containerID string) error
	Stop(ctx context.Context, containerID string) error
	Restart(ctx context.Context, containerID string) error
	Delete(ctx context.Context, containerID string) error
	Wait(ctx context.Context, containerID string) (int64, error)

	// Logging
	StreamLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error)

	// Exec operations
	ExecInteractive(ctx context.Context, containerID string, cmd []string, stdin io.Reader, stdout, stderr io.Writer, resize <-chan TerminalSize) error

	// Container management
	Create(ctx context.Context, config ContainerConfig) (string, error)
	List(ctx context.Context) ([]ContainerInfo, error)
	Inspect(ctx context.Context, containerID string) (*ContainerInfo, error)

	// Image operations
	PullImage(ctx context.Context, image string) error
	BuildImage(ctx context.Context, buildConfig BuildConfig) error
}

type ContainerConfig struct {
	Image         string
	Name          string
	Ports         map[string]string // host:container
	Environment   map[string]string
	Volumes       map[string]string // host:container
	Networks      []string
	NetworkMode   string // Network mode: "bridge", "host", "none", or container:<name|id>
	RestartPolicy string
	WorkingDir    string
	Command       []string
	Entrypoint    []string
	AutoRemove    bool              // Automatically remove container when it exits
	Privileged    bool              // Run container in privileged mode (needed for some build operations)
	Labels        map[string]string // Container labels for metadata and routing (e.g., Traefik)
}

type ContainerInfo struct {
	ID     string
	Name   string
	Image  string
	State  string
	Status string
	Ports  map[string]string
}

type TerminalSize struct {
	Height uint
	Width  uint
}

type BuildConfig struct {
	Context    string
	Dockerfile string
	Tag        string
	Args       map[string]string
	Target     string
}

type ManagerType string

const (
	Docker ManagerType = "docker"
	Podman ManagerType = "podman"
)

func NewContainerManager(managerType ManagerType) (ContainerManager, error) {
	switch managerType {
	case Docker:
		return NewDockerManager()
	case Podman:
		return NewPodmanManager()
	default:
		return nil, fmt.Errorf("unsupported container manager: %s", managerType)
	}
}
