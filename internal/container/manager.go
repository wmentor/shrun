package container

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

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

type ContainerStartSettings struct {
	Image     string
	Host      string
	Cmd       []string
	NetworkID string
	Envs      []string
}

func (mng *Manager) CreateAndStart(ctx context.Context, css ContainerStartSettings) (string, error) {
	baseConf := &container.Config{
		Hostname: css.Host,
		Image:    css.Image,
		Cmd:      strslice.StrSlice(css.Cmd),
		Env:      css.Envs,
	}

	hostConf := &container.HostConfig{
		Privileged: true,
	}

	resp, err := mng.client.ContainerCreate(ctx, baseConf, hostConf, nil, nil, css.Host)
	if err != nil {
		panic(err)
	}

	if err := mng.client.NetworkConnect(ctx, css.NetworkID, resp.ID, nil); err != nil {
		panic(err)
	}

	if err := mng.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return resp.ID, nil
}

func (mng *Manager) Exec(ctx context.Context, conID string, command string, username string) (int, error) {
	opts := types.ExecConfig{
		Tty:          true,
		Cmd:          []string{"sh", "-c", command},
		AttachStderr: true,
		AttachStdout: true,
		User:         username,
	}

	log.Println(command)

	eid, err := mng.client.ContainerExecCreate(ctx, conID, opts)
	if err != nil {
		log.Print(err)
		return 0, err
	}

	aresp, err := mng.client.ContainerExecAttach(ctx, eid.ID, types.ExecStartCheck{})
	if err != nil {
		return 0, err
	}
	defer aresp.Close()

	stdcopy.StdCopy(os.Stdout, os.Stderr, aresp.Reader)

	eresp, err := mng.client.ContainerExecInspect(ctx, eid.ID)
	if err != nil {
		return 0, err
	}

	return eresp.ExitCode, nil
}

func Shell(ctx context.Context, containerName string, username string) error {
	cmd := exec.CommandContext(ctx, "docker", "exec", "-ti", "-u", username, containerName, "/bin/bash")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (mng *Manager) StopAll(ctx context.Context) {
	contOpts := types.ContainerListOptions{All: true}

	containers, err := mng.client.ContainerList(ctx, contOpts)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	for _, container := range containers {
		if !mng.isOurContainer(container.Names) {
			continue
		}

		wg.Add(1)
		go func(id string, name string, state string) {
			defer wg.Done()
			if state == "running" {
				if err = mng.client.ContainerStop(ctx, id, nil); err != nil {
					log.Printf("stop container %s error: %v", name, err)
				} else {
					log.Printf("stop container %s success", name)
				}
			}
			opts := types.ContainerRemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			}

			if err = mng.client.ContainerRemove(ctx, id, opts); err != nil {
				log.Printf("remove container %s error: %v", name, err)
			} else {
				log.Printf("remove container %s success", name)
			}
		}(container.ID, container.Names[0], container.State)
	}

	wg.Wait()
}

func (mng *Manager) isOurContainer(names []string) bool {
	for _, name := range names {
		if strings.HasPrefix(name, "/"+common.GetObjectPrefix()) || strings.HasPrefix(name, common.GetObjectPrefix()) {
			return true
		}
	}
	return false
}
