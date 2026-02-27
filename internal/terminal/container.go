package terminal

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerManager struct {
	docker      *client.Client
	image       string
	memoryLimit int64
	cpuLimit    float64
}

func NewContainerManager(dockerHost, image, memoryLimit string, cpuLimit float64) (*ContainerManager, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	memBytes := parseMemoryLimit(memoryLimit)

	return &ContainerManager{
		docker:      cli,
		image:       image,
		memoryLimit: memBytes,
		cpuLimit:    cpuLimit,
	}, nil
}

type ContainerInfo struct {
	ID string
}

func (cm *ContainerManager) CreateContainer(ctx context.Context, projectName string) (*ContainerInfo, error) {
	cfg := &container.Config{
		Image:        cm.image,
		Cmd:          []string{"/bin/sh"},
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/workspace",
		Labels: map[string]string{
			"aihub.project": projectName,
			"aihub.type":    "sandbox",
		},
	}

	hostCfg := &container.HostConfig{
		Resources: container.Resources{
			Memory:   cm.memoryLimit,
			NanoCPUs: int64(cm.cpuLimit * 1e9),
		},
		SecurityOpt: []string{"no-new-privileges"},
		AutoRemove:  true,
	}

	resp, err := cm.docker.ContainerCreate(ctx, cfg, hostCfg, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := cm.docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return &ContainerInfo{ID: resp.ID}, nil
}

// AttachResult wraps the Docker attach response.
type AttachResult struct {
	Reader io.Reader
	Conn   io.WriteCloser
}

func (cm *ContainerManager) AttachContainer(ctx context.Context, containerID string) (*AttachResult, error) {
	resp, err := cm.docker.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}

	return &AttachResult{
		Reader: resp.Reader,
		Conn:   resp.Conn,
	}, nil
}

func (cm *ContainerManager) StopContainer(ctx context.Context, containerID string) error {
	timeout := 10
	stopOpts := container.StopOptions{Timeout: &timeout}
	if err := cm.docker.ContainerStop(ctx, containerID, stopOpts); err != nil {
		log.Printf("Failed to stop container %s: %v", containerID, err)
	}
	return nil
}

func (cm *ContainerManager) ResizeTerminal(ctx context.Context, containerID string, height, width uint) error {
	return cm.docker.ContainerResize(ctx, containerID, container.ResizeOptions{
		Height: height,
		Width:  width,
	})
}

func (cm *ContainerManager) Close() error {
	return cm.docker.Close()
}

func parseMemoryLimit(limit string) int64 {
	if len(limit) < 2 {
		return 256 * 1024 * 1024
	}
	suffix := limit[len(limit)-1]
	numStr := limit[:len(limit)-1]

	var num int64
	fmt.Sscanf(numStr, "%d", &num)

	switch suffix {
	case 'm', 'M':
		return num * 1024 * 1024
	case 'g', 'G':
		return num * 1024 * 1024 * 1024
	default:
		return 256 * 1024 * 1024
	}
}

// CleanupIdleSessions stops containers that have been idle for too long.
func (cm *ContainerManager) CleanupIdleSessions(ctx context.Context, containerIDs []string, _ time.Duration) {
	for _, id := range containerIDs {
		info, err := cm.docker.ContainerInspect(ctx, id)
		if err != nil {
			continue
		}
		if info.State != nil && info.State.Running {
			cm.StopContainer(ctx, id)
		}
	}
}
