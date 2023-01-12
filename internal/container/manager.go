package container

import (
	"context"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"

	"github.com/wmentor/shrun/internal/common"
)

type Manager struct {
	client *client.Client
}

func NewManager(client *client.Client) (*Manager, error) {
	if client == nil {
		return nil, common.ErrInvalidDockerClient
	}

	mng := &Manager{
		client: client,
	}

	return mng, nil
}

func (mng *Manager) CreateAndStart(ctx context.Context, imageName string, host string, cmd []string, netID string, envs []string) (string, error) {
	baseConf := &container.Config{
		Hostname: host,
		Image:    imageName,
		Cmd:      strslice.StrSlice(cmd),
		Env:      envs,
	}

	hostConf := &container.HostConfig{
		Privileged: true,
	}

	resp, err := mng.client.ContainerCreate(ctx, baseConf, hostConf, nil, nil, host)
	if err != nil {
		panic(err)
	}

	if err := mng.client.NetworkConnect(ctx, netID, resp.ID, nil); err != nil {
		panic(err)
	}

	if err := mng.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	mng.client.

	return resp.ID, nil
}

func (mng *Manager) StopAll(ctx context.Context) {
	contOpts := types.ContainerListOptions{All: true}

	containers, err := mng.client.ContainerList(ctx, contOpts)
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if !mng.isOurContainer(container.Names) {
			continue
		}
		if container.State == "running" {
			if err = mng.client.ContainerStop(ctx, container.ID, nil); err != nil {
				log.Printf("stop container %s error: %v", container.Names[0], err)
			} else {
				log.Printf("stop container %s success", container.Names[0])
			}
		}
		opts := types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}

		if err = mng.client.ContainerRemove(ctx, container.ID, opts); err != nil {
			log.Printf("remove container %s error: %v", container.Names[0], err)
		} else {
			log.Printf("remove container %s success", container.Names[0])
		}
	}
}

func (mng *Manager) isOurContainer(names []string) bool {
	for _, name := range names {
		if strings.HasPrefix(name, "/"+common.GetObjectPrefix()) || strings.HasPrefix(name, common.GetObjectPrefix()) {
			return true
		}
	}
	return false
}
