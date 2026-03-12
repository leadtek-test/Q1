package adapters

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/leadtek-test/q1/container/domain/container"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type ContainerRuntimeRepositoryDocker struct {
	client *client.Client
}

func NewContainerRuntimeRepositoryDocker() (*ContainerRuntimeRepositoryDocker, error) {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		return nil, err
	}
	return &ContainerRuntimeRepositoryDocker{client: cli}, nil
}

func (d ContainerRuntimeRepositoryDocker) Create(ctx context.Context, userID uint, spec container.CreateSpec, workspacePath string) (string, error) {
	if err := d.ensureImage(ctx, spec.Image); err != nil {
		return "", err
	}

	env := make([]string, 0, len(spec.Env))
	for k, v := range spec.Env {
		env = append(env, k+"="+v)
	}

	name := fmt.Sprintf("q1-u%s-%s-%s", strconv.FormatUint(uint64(userID), 10), sanitizeRuntimeName(spec.Name), uuid.NewString()[:8])
	resp, err := d.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &dockercontainer.Config{
			Image: spec.Image,
			Cmd:   spec.Command,
			Env:   env,
			Tty:   false,
		},
		HostConfig: &dockercontainer.HostConfig{
			Binds: []string{workspacePath + ":/workspace"},
		},
		Name: name,
	})
	if err != nil {
		return "", fmt.Errorf("docker create container failed: %w", err)
	}
	return resp.ID, nil
}

func (d ContainerRuntimeRepositoryDocker) Start(ctx context.Context, runtimeID string) error {
	_, err := d.client.ContainerStart(ctx, runtimeID, client.ContainerStartOptions{})
	return err
}

func (d ContainerRuntimeRepositoryDocker) Stop(ctx context.Context, runtimeID string) error {
	_, err := d.client.ContainerStop(ctx, runtimeID, client.ContainerStopOptions{})
	return err
}

func (d ContainerRuntimeRepositoryDocker) Delete(ctx context.Context, runtimeID string) error {
	_, err := d.client.ContainerRemove(ctx, runtimeID, client.ContainerRemoveOptions{Force: true})
	return err
}

func (d ContainerRuntimeRepositoryDocker) ensureImage(ctx context.Context, imageRef string) error {
	reader, err := d.client.ImagePull(ctx, imageRef, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("docker image pull failed: %w", err)
	}
	return reader.Wait(ctx)
}

func sanitizeRuntimeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "default"
	}
	var builder strings.Builder
	for _, char := range name {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		default:
			builder.WriteByte('-')
		}
	}
	output := strings.Trim(builder.String(), "-")
	if output == "" {
		return "default"
	}
	return output
}
