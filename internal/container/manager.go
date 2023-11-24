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
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"

	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/entities"
)

var (
	stopTimeout = 60
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
	exposedPorts, portBindings, err := nat.ParsePortSpecs(css.Ports)
	if err != nil {
		return "", err
	}

	baseConf := &container.Config{
		Hostname:     css.Host,
		Image:        css.Image,
		Cmd:          strslice.StrSlice(css.Cmd),
		ExposedPorts: exposedPorts,
		Env:          css.Envs,
	}

	hostConf := &container.HostConfig{
		Privileged:   true,
		PortBindings: portBindings,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: common.GetVolumeDir(),
				Target: "/mntdata",
			},
		},
		ShmSize: common.ShmSize,
	}

	if css.Debug {
		hostConf.SecurityOpt = []string{"apparmor=unconfined"}
		hostConf.CapAdd = []string{"SYS_PTRACE"}
	}

	if matched, _ := regexp.MatchString("n\\d+$", css.Host); matched {
		dataDir := filepath.Join(common.GetPgDataDir(), css.Host)
		os.RemoveAll(dataDir)
		if css.MountData {
			os.Mkdir(dataDir, common.AccessPerm)
			hostConf.Mounts = append(hostConf.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: dataDir,
				Target: fmt.Sprintf("/var/lib/pgpro/sdm-%d/data", common.PgVersion),
			})
		}
	}

	if strings.HasSuffix(css.Host, "prometheus") {
		dataDir := filepath.Join(common.GetDataDir(), "prometheus")
		os.RemoveAll(dataDir)
		os.Mkdir(dataDir, common.AccessPerm)
		hostConf.Mounts = append(hostConf.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: dataDir,
			Target: "/prometheus/wal",
		})
	}

	if strings.HasSuffix(css.Host, "cadvisor") {
		hostConf.Mounts = append(hostConf.Mounts,
			mount.Mount{
				Type:     mount.TypeBind,
				Source:   "/",
				Target:   "/rootfs",
				ReadOnly: true,
			},
			mount.Mount{
				Type:   mount.TypeBind,
				Source: "/var/run",
				Target: "/var/run",
			},
			mount.Mount{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
			mount.Mount{
				Type:     mount.TypeBind,
				Source:   "/sys",
				Target:   "/sys",
				ReadOnly: true,
			},
			mount.Mount{
				Type:     mount.TypeBind,
				Source:   "/var/lib/docker",
				Target:   "/var/lib/docker",
				ReadOnly: true,
			},
		)

		hostConf.Devices = []container.DeviceMapping{{PathOnHost: "/dev/kmsg"}}
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

	if css.Host == "core" {
		hostConf.Mounts = append(hostConf.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: common.GetDataDir(),
			Target: "/repo",
		})
	}

	hostConf.Memory = css.MemoryLimit
	if css.CPU != 0 {
		hostConf.NanoCPUs = int64(1000000000 * css.CPU)
	}

	var nc *network.NetworkingConfig = nil

	if css.NetworkID != "" {
		nc = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				css.NetworkID: {NetworkID: css.NetworkID},
			},
		}
	}

	resp, err := mng.client.ContainerCreate(ctx, baseConf, hostConf, nc, nil, css.Host)
	if err != nil {
		log.Printf("create container %s error: %v", css.Host, err)
		return "", err
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

func (mng *Manager) ShellCommand(ctx context.Context, containerName string, username string, workDir string, command []string) error {
	args := []string{"exec", "-ti", "-u", username}
	if workDir != "" {
		args = append(args, "-w", workDir)
	}
	args = append(args, containerName)
	args = append(args, command...)
	cmd := exec.CommandContext(ctx, "docker", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (mng *Manager) Renew(ctx context.Context) error {
	list1, list2 := mng.makeContList(ctx)

	if err := mng.renewList(ctx, list2); err != nil {
		return err
	}

	return mng.renewList(ctx, list1)
}

func (mng *Manager) renewList(ctx context.Context, containers []types.Container) error {
	var wg sync.WaitGroup

	errCh := make(chan error, len(containers))

	for _, container := range containers {
		if !mng.isOurContainer(container.Names) {
			continue
		}

		wg.Add(1)
		go func(id string, name string, state string) {
			defer wg.Done()
			if state != "running" {
				if err := mng.client.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
					log.Printf("start container %s error: %v", name, err)
					return
				}
				log.Printf("start container %s", strings.TrimLeft(name, "/"))
			}
		}(container.ID, container.Names[0], container.State)
	}

	wg.Wait()

	close(errCh)

	for e := range errCh {
		return e
	}

	return nil
}

func (mng *Manager) StopAll(ctx context.Context, remove bool) {
	list1, list2 := mng.makeContList(ctx)

	mng.stopList(ctx, list1, remove)
	mng.stopList(ctx, list2, remove)
}

func (mng *Manager) makeContList(ctx context.Context) ([]types.Container, []types.Container) {
	contOpts := types.ContainerListOptions{All: true}

	containers, err := mng.client.ContainerList(ctx, contOpts)
	if err != nil {
		panic(err)
	}

	prefix := common.GetObjectPrefix() + "e"

	list1 := make([]types.Container, 0, len(containers))
	list2 := make([]types.Container, 0, len(containers))

	for _, con := range containers {
		if strings.HasPrefix(con.Names[0], prefix) || strings.HasPrefix(con.Names[0], "/"+prefix) {
			list2 = append(list2, con)
		} else {
			list1 = append(list1, con)
		}
	}

	return list1, list2
}

func (mng *Manager) stopList(ctx context.Context, containers []types.Container, remove bool) {
	var wg sync.WaitGroup

	for _, curContainer := range containers {
		if !mng.isOurContainer(curContainer.Names) {
			continue
		}

		wg.Add(1)
		go func(id string, name string, state string) {
			defer wg.Done()
			if state == "running" {
				var opts container.StopOptions
				if err := mng.client.ContainerStop(ctx, id, opts); err != nil {
					log.Printf("stop container %s error: %v", name, err)
				} else {
					log.Printf("stop container %s success", name)
				}
			}
			if remove {
				opts := types.ContainerRemoveOptions{
					RemoveVolumes: true,
					Force:         true,
				}

				if err := mng.client.ContainerRemove(ctx, id, opts); err != nil {
					log.Printf("remove container %s error: %v", name, err)
				} else {
					log.Printf("remove container %s success", name)
				}
				mng.removePgData(name)
			}
		}(curContainer.ID, curContainer.Names[0], curContainer.State)
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
	curContainer, err := mng.GetContainer(ctx, name)
	if err != nil {
		if !errors.Is(err, common.ErrNotFound) {
			return err
		}
		return nil
	}

	if curContainer.Status == "running" {
		var opts container.StopOptions
		if err = mng.client.ContainerStop(ctx, curContainer.ID, opts); err != nil {
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

	if err = mng.client.ContainerRemove(ctx, curContainer.ID, opts); err != nil {
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

func (mng *Manager) GetExporterName(num int) string {
	return fmt.Sprintf("%s%s%d", common.GetObjectPrefix(), "pe", num)
}

func (mng *Manager) StartPrometheusExporter(ctx context.Context, num int, netID string) error {
	hostname := mng.GetExporterName(num)
	log.Printf("start %s", hostname)

	envs := common.GetEnvs()

	envs = append(envs,
		"DATA_SOURCE_NAME=postgresql://postgres:12345@"+common.GetNodeName(num)+":5432/postgres?sslmode=disable",
		"PG_EXPORTER_DISABLE_DEFAULT_METRICS=true",
		"PG_EXPORTER_DISABLE_SETTINGS_METRICS=true",
	)

	opts := entities.ContainerStartSettings{
		Image:     "prometheuscommunity/postgres-exporter",
		Host:      hostname,
		NetworkID: netID,
		Envs:      envs,
	}

	if _, err := mng.CreateAndStart(ctx, opts); err != nil {
		return err
	}

	return nil
}
