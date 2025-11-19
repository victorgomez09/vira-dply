package manager

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerManager struct {
	client *client.Client
}

func NewDockerManager() (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerManager{client: cli}, nil
}

func (d *DockerManager) Start(ctx context.Context, containerID string) error {
	return d.client.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (d *DockerManager) Stop(ctx context.Context, containerID string) error {
	return d.client.ContainerStop(ctx, containerID, container.StopOptions{})
}

func (d *DockerManager) Restart(ctx context.Context, containerID string) error {
	return d.client.ContainerRestart(ctx, containerID, container.StopOptions{})
}

func (d *DockerManager) Delete(ctx context.Context, containerID string) error {
	return d.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}

func (d *DockerManager) Wait(ctx context.Context, containerID string) (int64, error) {
	statusCh, errCh := d.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return -1, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		return status.StatusCode, nil
	}
	return 0, nil
}

func (d *DockerManager) StreamLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	return d.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
	})
}

func (d *DockerManager) Create(ctx context.Context, config ContainerConfig) (string, error) {
	// Convert port mappings
	portBindings := make(nat.PortMap)
	exposedPorts := make(nat.PortSet)

	for hostPort, containerPort := range config.Ports {
		port := nat.Port(containerPort + "/tcp")
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{HostPort: hostPort},
		}
	}

	// Convert environment variables
	env := make([]string, 0, len(config.Environment))
	for key, value := range config.Environment {
		env = append(env, key+"="+value)
	}

	// Convert volumes
	binds := make([]string, 0, len(config.Volumes))
	for hostPath, containerPath := range config.Volumes {
		binds = append(binds, hostPath+":"+containerPath)
	}

	containerConfig := &container.Config{
		Image:        config.Image,
		Env:          env,
		ExposedPorts: exposedPorts,
		WorkingDir:   config.WorkingDir,
		Cmd:          config.Command,
		Entrypoint:   config.Entrypoint,
		Labels:       config.Labels,
	}

	hostConfig := &container.HostConfig{
		PortBindings:  portBindings,
		Binds:         binds,
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyMode(config.RestartPolicy)},
		AutoRemove:    config.AutoRemove,
		Privileged:    config.Privileged,
		NetworkMode:   container.NetworkMode(config.NetworkMode),
	}

	networkConfig := &network.NetworkingConfig{}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, config.Name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	for _, networkName := range config.Networks {
		if err := d.client.NetworkConnect(ctx, networkName, resp.ID, nil); err != nil {
			d.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
			return "", fmt.Errorf("failed to connect container to network %s: %w", networkName, err)
		}
	}

	return resp.ID, nil
}

func (d *DockerManager) List(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]ContainerInfo, len(containers))
	for i, c := range containers {
		ports := make(map[string]string)
		for _, port := range c.Ports {
			if port.PublicPort > 0 {
				ports[fmt.Sprintf("%d", port.PublicPort)] = fmt.Sprintf("%d", port.PrivatePort)
			}
		}

		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
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

func (d *DockerManager) Inspect(ctx context.Context, containerID string) (*ContainerInfo, error) {
	inspect, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	ports := make(map[string]string)
	if inspect.NetworkSettings != nil {
		for port, bindings := range inspect.NetworkSettings.Ports {
			if len(bindings) > 0 && bindings[0].HostPort != "" {
				ports[bindings[0].HostPort] = string(port)
			}
		}
	}

	return &ContainerInfo{
		ID:     inspect.ID,
		Name:   strings.TrimPrefix(inspect.Name, "/"),
		Image:  inspect.Config.Image,
		State:  inspect.State.Status,
		Status: inspect.State.Status,
		Ports:  ports,
	}, nil
}

func (d *DockerManager) PullImage(ctx context.Context, imageName string) error {
	reader, err := d.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read the response to ensure the pull completes
	_, err = io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read pull response: %w", err)
	}

	return nil
}

func (d *DockerManager) BuildImage(ctx context.Context, buildConfig BuildConfig) error {
	// Implementation would use docker build API
	// For now, this is a placeholder
	return fmt.Errorf("docker build not yet implemented")
}

func (d *DockerManager) ExecInteractive(ctx context.Context, containerID string, cmd []string, stdin io.Reader, stdout, stderr io.Writer, resize <-chan TerminalSize) error {
	execConfig := container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          cmd,
	}

	execIDResp, err := d.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := d.client.ContainerExecAttach(ctx, execIDResp.ID, container.ExecStartOptions{
		Tty: true,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer resp.Close()

	errChan := make(chan error, 3)

	go func() {
		_, err := io.Copy(resp.Conn, stdin)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(stdout, resp.Reader)
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
				if err := d.client.ContainerExecResize(ctx, execIDResp.ID, container.ResizeOptions{
					Height: size.Height,
					Width:  size.Width,
				}); err != nil {
					errChan <- fmt.Errorf("failed to resize terminal: %w", err)
					return
				}
			}
		}
	}()

	return <-errChan
}
