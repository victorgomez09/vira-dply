package manager

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/containers/common/libnetwork/types"
	"github.com/containers/podman/v5/pkg/api/handlers"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/containers/podman/v5/pkg/specgen"
	"github.com/docker/docker/api/types/container"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type PodmanManager struct {
	socketPath string
	connCtx    context.Context
}

func NewPodmanManager() (*PodmanManager, error) {
	// Get user ID for socket path
	uid := os.Getuid()
	socketPath := fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", uid)

	// If running as root, use system socket
	if uid == 0 {
		socketPath = "unix:///run/podman/podman.sock"
	}

	// Parse socket URL
	conn, err := url.Parse(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse socket path: %w", err)
	}

	// Create connection context
	connCtx, err := bindings.NewConnection(context.Background(), conn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Podman socket: %w", err)
	}

	return &PodmanManager{
		socketPath: socketPath,
		connCtx:    connCtx,
	}, nil
}

func (p *PodmanManager) Start(ctx context.Context, containerID string) error {
	err := containers.Start(p.connCtx, containerID, nil)
	if err != nil {
		return fmt.Errorf("failed to start container %s: %w", containerID, err)
	}
	return nil
}

func (p *PodmanManager) Stop(ctx context.Context, containerID string) error {
	timeout := uint(10) // 10 seconds timeout
	options := &containers.StopOptions{
		Timeout: &timeout,
	}
	err := containers.Stop(p.connCtx, containerID, options)
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}
	return nil
}

func (p *PodmanManager) Restart(ctx context.Context, containerID string) error {
	timeout := 10
	options := &containers.RestartOptions{
		Timeout: &timeout,
	}
	err := containers.Restart(p.connCtx, containerID, options)
	if err != nil {
		return fmt.Errorf("failed to restart container %s: %w", containerID, err)
	}
	return nil
}

func (p *PodmanManager) Delete(ctx context.Context, containerID string) error {
	force := true
	options := &containers.RemoveOptions{
		Force: &force,
	}
	reports, err := containers.Remove(p.connCtx, containerID, options)
	if err != nil {
		return fmt.Errorf("failed to delete container %s: %w", containerID, err)
	}

	// Check for any errors in the removal reports
	for _, report := range reports {
		if report.Err != nil {
			return fmt.Errorf("error removing container %s: %w", containerID, report.Err)
		}
	}

	return nil
}

func (p *PodmanManager) Wait(ctx context.Context, containerID string) (int64, error) {
	exitCode, err := containers.Wait(p.connCtx, containerID, nil)
	if err != nil {
		return -1, fmt.Errorf("error waiting for container: %w", err)
	}
	return int64(exitCode), nil
}

func (p *PodmanManager) StreamLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	timestamps := true
	since := ""
	until := ""
	tail := -1

	stdoutChan := make(chan string)
	stderrChan := make(chan string)

	// Use the correct v5 Logs signature with 5 parameters
	options := &containers.LogOptions{}
	options.WithFollow(follow)
	options.WithTimestamps(timestamps)
	options.WithSince(since)
	options.WithUntil(until)
	options.WithTail(strconv.Itoa(tail))

	// Create a pipe to convert the channels to ReadCloser
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		defer close(stdoutChan)
		defer close(stderrChan)

		// Start the logs streaming
		err := containers.Logs(p.connCtx, containerID, options, stdoutChan, stderrChan)
		if err != nil {
			return
		}
	}()

	go func() {
		defer pw.Close()
		for {
			select {
			case log, ok := <-stdoutChan:
				if !ok {
					return
				}
				if _, err := pw.Write([]byte(log)); err != nil {
					return
				}
			case log, ok := <-stderrChan:
				if !ok {
					return
				}
				if _, err := pw.Write([]byte(log)); err != nil {
					return
				}
			}
		}
	}()

	return pr, nil
}

func (p *PodmanManager) Create(ctx context.Context, config ContainerConfig) (string, error) {
	spec := &specgen.SpecGenerator{
		ContainerBasicConfig: specgen.ContainerBasicConfig{
			Name:   config.Name,
			Remove: &config.AutoRemove,
			Labels: config.Labels,
		},
		ContainerStorageConfig: specgen.ContainerStorageConfig{
			Image: config.Image,
		},
		ContainerSecurityConfig: specgen.ContainerSecurityConfig{
			Privileged: &config.Privileged,
		},
	}

	// Set working directory
	if config.WorkingDir != "" {
		spec.WorkDir = config.WorkingDir
	}

	// Set restart policy
	if config.RestartPolicy != "" {
		spec.RestartPolicy = config.RestartPolicy
	}

	// Set entrypoint and command
	if len(config.Entrypoint) > 0 {
		spec.Entrypoint = config.Entrypoint
	}
	if len(config.Command) > 0 {
		spec.Command = config.Command
	}

	// Set environment variables
	if len(config.Environment) > 0 {
		env := make(map[string]string)
		for k, v := range config.Environment {
			env[k] = v
		}
		spec.Env = env
	}

	// Set port mappings
	if len(config.Ports) > 0 {
		portMappings := make([]types.PortMapping, 0, len(config.Ports))
		for hostPort, containerPort := range config.Ports {
			// Handle port ranges and protocol parsing
			hostPortNum, hostProtocol, err := parsePortSpec(hostPort)
			if err != nil {
				return "", fmt.Errorf("invalid host port %s: %w", hostPort, err)
			}
			containerPortNum, containerProtocol, err := parsePortSpec(containerPort)
			if err != nil {
				return "", fmt.Errorf("invalid container port %s: %w", containerPort, err)
			}

			// Use the protocol from container port, fallback to host port protocol, default to tcp
			protocol := "tcp"
			if containerProtocol != "" {
				protocol = containerProtocol
			} else if hostProtocol != "" {
				protocol = hostProtocol
			}

			portMappings = append(portMappings, types.PortMapping{
				HostPort:      uint16(hostPortNum),
				ContainerPort: uint16(containerPortNum),
				Protocol:      protocol,
			})
		}
		spec.PortMappings = portMappings
	}

	// Set volume mounts
	if len(config.Volumes) > 0 {
		mounts := make([]specs.Mount, 0, len(config.Volumes))
		for hostPath, containerPath := range config.Volumes {
			mount := specs.Mount{
				Source:      hostPath,
				Destination: containerPath,
				Type:        "bind",
				Options:     []string{"rbind"},
			}

			// Check for read-only flag
			if strings.HasSuffix(containerPath, ":ro") {
				mount.Destination = strings.TrimSuffix(containerPath, ":ro")
				mount.Options = append(mount.Options, "ro")
			}

			mounts = append(mounts, mount)
		}
		spec.Mounts = mounts
	}

	// Set network configuration
	if len(config.Networks) > 0 {
		spec.Networks = make(map[string]types.PerNetworkOptions)
		for _, network := range config.Networks {
			spec.Networks[network] = types.PerNetworkOptions{}
		}
	}

	// Create container
	createResponse, err := containers.CreateWithSpec(p.connCtx, spec, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return createResponse.ID, nil
}

// Helper function to parse port specifications
func parsePortSpec(portSpec string) (uint16, string, error) {
	parts := strings.Split(portSpec, "/")
	portStr := parts[0]
	protocol := ""

	if len(parts) > 1 {
		protocol = parts[1]
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, "", err
	}

	return uint16(port), protocol, nil
}

func (p *PodmanManager) List(ctx context.Context) ([]ContainerInfo, error) {
	all := true
	options := &containers.ListOptions{
		All: &all,
	}

	containerList, err := containers.List(p.connCtx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]ContainerInfo, len(containerList))
	for i, c := range containerList {
		ports := make(map[string]string)
		for _, port := range c.Ports {
			if port.HostPort > 0 {
				hostPortStr := strconv.FormatUint(uint64(port.HostPort), 10)
				containerPortStr := strconv.FormatUint(uint64(port.ContainerPort), 10)

				// Include protocol if not tcp
				if port.Protocol != "tcp" {
					containerPortStr += "/" + port.Protocol
				}

				ports[hostPortStr] = containerPortStr
			}
		}

		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			// Remove leading slash if present
			if strings.HasPrefix(name, "/") {
				name = name[1:]
			}
		}

		result[i] = ContainerInfo{
			ID:     c.ID,
			Name:   name,
			Image:  c.Image,
			State:  c.State,
			Status: c.Status,
			Ports:  ports,
		}
	}

	return result, nil
}

func (p *PodmanManager) Inspect(ctx context.Context, containerID string) (*ContainerInfo, error) {
	size := false
	options := &containers.InspectOptions{
		Size: &size,
	}

	inspectData, err := containers.Inspect(p.connCtx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Extract port mappings with improved parsing
	ports := make(map[string]string)
	if inspectData.NetworkSettings != nil && inspectData.NetworkSettings.Ports != nil {
		for containerPortProto, bindings := range inspectData.NetworkSettings.Ports {
			if len(bindings) > 0 && bindings[0].HostPort != "" {
				// Parse container port and protocol
				parts := strings.Split(containerPortProto, "/")
				containerPort := parts[0]
				protocol := "tcp"
				if len(parts) > 1 {
					protocol = parts[1]
				}

				// Format the container port with protocol if not tcp
				containerPortFormatted := containerPort
				if protocol != "tcp" {
					containerPortFormatted += "/" + protocol
				}

				ports[bindings[0].HostPort] = containerPortFormatted
			}
		}
	}

	name := inspectData.Name
	if strings.HasPrefix(name, "/") {
		name = name[1:] // Remove leading slash
	}

	return &ContainerInfo{
		ID:     inspectData.ID,
		Name:   name,
		Image:  inspectData.Config.Image,
		State:  inspectData.State.Status,
		Status: inspectData.State.Status,
		Ports:  ports,
	}, nil
}

func (p *PodmanManager) PullImage(ctx context.Context, image string) error {
	options := &images.PullOptions{}
	_, err := images.Pull(p.connCtx, image, options)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", image, err)
	}

	return nil
}

func (p *PodmanManager) BuildImage(ctx context.Context, buildConfig BuildConfig) error {
	// Prepare build options with v5 improvements
	// buildOptions := images.BuildOptions{
	// 	Jobs: func() *int { j := 1; return &j }(), // Parallel build jobs
	// }

	// // Set dockerfile
	// if buildConfig.Dockerfile != "" {
	// 	buildOptions.ContainerFiles = buildConfig.Dockerfile
	// }

	// // Set build args
	// if len(buildConfig.Args) > 0 {
	// 	buildOptions.Args = make(map[string]string)
	// 	for k, v := range buildConfig.Args {
	// 		buildOptions.Args[k] = v
	// 	}
	// }

	// // Set target
	// if buildConfig.Target != "" {
	// 	buildOptions.Target = buildConfig.Target
	// }

	// // Set tags
	// if buildConfig.Tag != "" {
	// 	buildOptions.Tags = []string{buildConfig.Tag}
	// }

	// // Determine context path
	// contextPath := "."
	// if buildConfig.Context != "" {
	// 	contextPath = buildConfig.Context
	// }

	// // Convert to absolute path
	// absPath, err := filepath.Abs(contextPath)
	// if err != nil {
	// 	return fmt.Errorf("failed to get absolute path: %w", err)
	// }

	// // Build the image
	// buildReport, err := images.Build(p.connCtx, []string{absPath}, buildOptions)
	// if err != nil {
	// 	return fmt.Errorf("failed to build image: %w", err)
	// }

	// // Check build report for errors
	// if buildReport.Error != "" {
	// 	return fmt.Errorf("build failed: %s", buildReport.Error)
	// }

	// return nil
	return fmt.Errorf("docker build not yet implemented")
}

// Additional v5 methods for enhanced functionality

// GetStats returns container statistics
func (p *PodmanManager) GetStats(ctx context.Context, containerID string) (*entities.ContainerStatsReport, error) {
	stream := false
	options := &containers.StatsOptions{
		Stream: &stream,
	}

	statsChan, err := containers.Stats(p.connCtx, []string{containerID}, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Get the first (and only) stats report
	for stats := range statsChan {
		if stats.Error != nil {
			return nil, fmt.Errorf("stats error: %w", stats.Error)
		}
		return &stats, nil
	}

	return nil, fmt.Errorf("no stats received")
}

// Exec runs a command in a container
func (p *PodmanManager) Exec(ctx context.Context, containerID string, cmd []string) error {
	createConfig := &handlers.ExecCreateConfig{
		ExecOptions: container.ExecOptions{
			Cmd: cmd,
			Tty: false,
		},
	}

	sessionID, err := containers.ExecCreate(p.connCtx, containerID, createConfig)

	if err != nil {
		return fmt.Errorf("failed to create exec session: %w", err)
	}

	err = containers.ExecStart(p.connCtx, sessionID, nil)
	if err != nil {
		return fmt.Errorf("failed to start exec session: %w", err)
	}

	return nil
}

func (p *PodmanManager) ExecInteractive(ctx context.Context, containerID string, cmd []string, stdin io.Reader, stdout, stderr io.Writer, resize <-chan TerminalSize) error {
	createConfig := &handlers.ExecCreateConfig{
		ExecOptions: container.ExecOptions{
			Cmd:          cmd,
			Tty:          true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
		},
	}

	sessionID, err := containers.ExecCreate(p.connCtx, containerID, createConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec session: %w", err)
	}

	startOptions := &containers.ExecStartAndAttachOptions{}

	errChan := make(chan error, 2)

	go func() {
		err := containers.ExecStartAndAttach(p.connCtx, sessionID, startOptions)
		errChan <- err
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case size, ok := <-resize:
				if !ok {
					return
				}
				height := int(size.Height)
				width := int(size.Width)
				resizeOptions := &containers.ResizeExecTTYOptions{}
				resizeOptions.WithHeight(height)
				resizeOptions.WithWidth(width)
				if err := containers.ResizeExecTTY(p.connCtx, sessionID, resizeOptions); err != nil {
					errChan <- fmt.Errorf("failed to resize terminal: %w", err)
					return
				}
			}
		}
	}()

	return <-errChan
}

// Helper method to close the connection
func (p *PodmanManager) Close() error {
	// The bindings don't provide an explicit close method
	// The connection will be closed when the context is cancelled
	return nil
}
