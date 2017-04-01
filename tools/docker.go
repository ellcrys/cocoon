package tools

import (
	"fmt"
	"time"

	"context"
	"strings"

	docker "github.com/ncodes/go-dockerclient"
)

var (
	// DockerEndpoint is the docker engine API endpoint
	DockerEndpoint = "unix:///var/run/docker.sock"

	// ErrContainerNotFound indicates a non-existent container
	ErrContainerNotFound = fmt.Errorf("container not found")
)

// DeleteContainer deletes a container and optionally
// deletes its image, remove the volumes or force delete if running.
func DeleteContainer(id string, removeImage, leaveVolume, noForceDelete bool) error {

	client, err := docker.NewClient(DockerEndpoint)
	if err != nil {
		return err
	}

	container, err := client.InspectContainer(id)
	if err != nil {
		if strings.Contains(err.Error(), "No such container:") {
			return ErrContainerNotFound
		}
		return err
	}

	image := container.Config.Image

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: !leaveVolume,
		Force:         !noForceDelete,
		Context:       ctx,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %s", err)
	}

	if removeImage {
		if err = client.RemoveImage(image); err != nil {
			return fmt.Errorf("failed to remove image: %s", err)
		}
	}

	return nil
}
