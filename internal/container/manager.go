package container

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/entities"
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

func (mng *Manager) CreateAndStart(ctx context.Context, css entities.ContainerStartSettings) (string, error) {
	baseConf := &container.Config{
		Hostname: css.Host,
		Image:    css.Image,
		Cmd:      strslice.StrSlice(css.Cmd),
		Env:      css.Envs,
	}

	hostConf := &container.HostConfig{
		Privileged: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: common.GetVolumeDir(),
				Target: "/mntdata",
			},
		},
	}

	if matched, _ := regexp.MatchString("n\\d+$", css.Host); matched {
		dataDir := filepath.Join(common.GetPgDataDir(), css.Host)
		os.RemoveAll(dataDir)
		if css.MountData {
			os.Mkdir(dataDir, 0755)
			hostConf.Mounts = append(hostConf.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: dataDir,
				Target: fmt.Sprintf("/var/lib/pgpro/sdm-%d/data", common.PgVersion),
			})
		}
	}

	if css.Host == "gobuilder" {
		hostConf.Mounts = append(hostConf.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: common.GetDataDir(),
			Target: "/repo",
		})

		hostConf.Mounts = append(hostConf.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: common.GetGoModDir(),
			Target: "/go/pkg",
		})
	}

	hostConf.Memory = css.MemoryLimit
	if css.CPU != 0 {
		hostConf.NanoCPUs = int64(1000000000 * css.CPU)
	}

	resp, err := mng.client.ContainerCreate(ctx, baseConf, hostConf, nil, nil, css.Host)
	if err != nil {
		log.Printf("create container %s error: %v", css.Host, err)
		return "", err
	}

	if css.NetworkID != "" {
		if err := mng.client.NetworkConnect(ctx, css.NetworkID, resp.ID, nil); err != nil {
			log.Printf("netword connect failed: %v", err)
			return "", err
		}
	}

	if err := mng.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Printf("start container %s error: %v", css.Host, err)
		return "", err
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

func (mng *Manager) ShellCommand(ctx context.Context, containerName string, username string, command []string) error {
	args := []string{"exec", "-ti", "-u", username, containerName}
	args = append(args, command...)
	cmd := exec.CommandContext(ctx, "docker", args...)

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
			mng.removePgData(name)
		}(container.ID, container.Names[0], container.State)
	}

	wg.Wait()
}

func (mng *Manager) GetContainer(ctx context.Context, name string) (entities.Container, error) {
	var result entities.Container

	contOpts := types.ContainerListOptions{All: true}

	list, err := mng.client.ContainerList(ctx, contOpts)
	if err != nil {
		return result, err
	}

	for _, item := range list {
		if mng.checkName(item.Names, name) {
			result.ID = item.ID
			result.Names = item.Names
			result.Status = item.State
			return result, nil
		}
	}

	return result, common.ErrNotFound
}

func (mng *Manager) RemoveContainer(ctx context.Context, name string) error {
	container, err := mng.GetContainer(ctx, name)
	if err != nil {
		if !errors.Is(err, common.ErrNotFound) {
			return err
		}
		return nil
	}

	if container.Status == "running" {
		if err = mng.client.ContainerStop(ctx, container.ID, nil); err != nil {
			log.Printf("stop container %s error: %v", name, err)
			return err
		} else {
			log.Printf("stop container %s success", name)
		}
	}

	opts := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err = mng.client.ContainerRemove(ctx, container.ID, opts); err != nil {
		log.Printf("remove container %s error: %v", name, err)
		return err
	} else {
		log.Printf("remove container %s success", name)
	}

	mng.removePgData(name)

	return nil
}

func (mng *Manager) removePgData(name string) {
	name = strings.TrimLeft(name, "/")
	dataDir := filepath.Join(common.GetPgDataDir(), name)
	os.RemoveAll(dataDir)
}

func (mng *Manager) isOurContainer(names []string) bool {
	for _, name := range names {
		if strings.HasPrefix(name, "/"+common.GetObjectPrefix()) || strings.HasPrefix(name, common.GetObjectPrefix()) {
			return true
		}
	}
	return false
}

func (mng *Manager) checkName(names []string, searchName string) bool {
	for _, name := range names {
		if name == "/"+searchName || name == searchName {
			return true
		}
	}

	return false
}
