package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/build"
	"github.com/wmentor/shrun/internal/cases/stop"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/entities"
	"github.com/wmentor/shrun/internal/image"
	"github.com/wmentor/shrun/internal/network"
)

var (
	_ cmd.CobraCommand = (*CommandBuild)(nil)
)

type CommandStart struct {
	command       *cobra.Command
	cli           *client.Client
	nodesCount    int
	skipNodeAdd   bool
	force         bool
	updateDockers bool
	memoryLimit   string
	cpuLimit      float64
	maxIOps       int64
}

func NewCommandStart(cli *client.Client) *CommandStart {
	c := &CommandStart{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "start",
		Short: "start cluster",
		RunE:  c.exec,
	}

	cc.Flags().IntVarP(&c.nodesCount, "nodes", "n", 2, "nodes count")
	cc.Flags().BoolVar(&c.skipNodeAdd, "skip-node-add", false, "skip shardmanctl nodes add")
	cc.Flags().BoolVarP(&c.force, "force", "f", false, "force start. (if already started then restart)")
	cc.Flags().BoolVarP(&c.updateDockers, "update", "u", false, "rebuild utils and update dockers")
	cc.Flags().StringVar(&c.memoryLimit, "memory", "", "memory limit")
	//cc.Flags().Int64Var(&c.maxIOps, "iops", 0, "IOps limit")
	cc.Flags().Float64Var(&c.cpuLimit, "cpu", 0, "cpu limit")

	c.command = cc

	return c
}

func (c *CommandStart) Command() *cobra.Command {
	return c.command
}

func (c *CommandStart) exec(cc *cobra.Command, _ []string) error {
	if c.nodesCount < 1 {
		return errors.New("invalid nodes count")
	}

	ctx := cc.Context()

	networker, err := network.NewManager(c.cli)
	if err != nil {
		return err
	}

	manager, err := container.NewManager(c.cli)
	if err != nil {
		return err
	}

	started, err := networker.CheckNetworkExists(ctx)
	if err != nil {
		return err
	}

	if started {
		if err = c.isStarted(ctx, manager, networker); err != nil {
			return err
		}
	}

	if c.updateDockers {
		if err = c.update(ctx); err != nil {
			return err
		}
	}

	netID, err := networker.CreateNetwork(ctx)
	if err != nil {
		return fmt.Errorf("create network error: %w", err)
	}

	etcdList, err := common.GetEtcdList()
	if err != nil {
		return err
	}

	etcdClusterMaker := strings.Builder{}
	for i := range etcdList {
		hostname := fmt.Sprintf("%se%d", common.GetObjectPrefix(), i+1)
		if i > 0 {
			etcdClusterMaker.WriteRune(',')
		}
		etcdClusterMaker.WriteString(hostname)
		etcdClusterMaker.WriteRune('=')
		etcdClusterMaker.WriteString(fmt.Sprintf("http://%s:2380", hostname))
	}

	containerIDs := map[string]string{}

	for i := range etcdList {
		hostname := fmt.Sprintf("%se%d", common.GetObjectPrefix(), i+1)

		cmdParams := []string{
			"/usr/local/bin/etcd",
			fmt.Sprintf("--name=%s", hostname),
			fmt.Sprintf("--initial-advertise-peer-urls=http://%s:2380", hostname),
			"--listen-peer-urls=http://0.0.0.0:2380",
			fmt.Sprintf("--advertise-client-urls=http://%s:2379", hostname),
			"--listen-client-urls=http://0.0.0.0:2379",
			fmt.Sprintf("--initial-cluster=%s", etcdClusterMaker.String()),
		}

		log.Printf("start %s", hostname)

		opts := entities.ContainerStartSettings{
			Image:     "quay.io/coreos/etcd:v3.5.6",
			Host:      hostname,
			Cmd:       cmdParams,
			NetworkID: netID,
		}

		if id, err := manager.CreateAndStart(ctx, opts); err == nil {
			containerIDs[hostname] = id
		} else {
			log.Println("failed")
		}
	}

	clusterName, _ := common.GetClusterName()
	logLevel, _ := common.GetLogLevel()

	envs := []string{
		fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
		fmt.Sprintf("SDM_CLUSTER_NAME=%s", clusterName),
		fmt.Sprintf("SDM_LOG_LEVEL=%s", logLevel),
		fmt.Sprintf("SDM_STORE_ENDPOINTS=%s", strings.Join(etcdList, ",")),
	}

	for i := 0; i < c.nodesCount; i++ {
		hostname := common.GetNodeName(i + 1)
		log.Printf("start %s", hostname)

		opts := entities.ContainerStartSettings{
			Image:     "shardman:latest",
			Host:      hostname,
			NetworkID: netID,
			Envs:      envs,
		}

		if c.memoryLimit != "" {
			size, err := units.RAMInBytes(c.memoryLimit)
			if err != nil {
				return fmt.Errorf("invalid memory limit: %v", c.memoryLimit)
			}
			opts.MemoryLimit = size
		}

		if c.cpuLimit != 0 {
			opts.CPU = c.cpuLimit
		}

		opts.MaxIOps = c.maxIOps

		if id, err := manager.CreateAndStart(ctx, opts); err == nil {
			containerIDs[hostname] = id
		} else {
			log.Println("failed")
			return err
		}
	}

	node1ID := containerIDs[common.GetNodeName(1)]

	code, err := manager.Exec(ctx, node1ID, "shardmanctl init -f /etc/shardman/sdmspec.json", "postgres")
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("command status code: %d", code)
	}

	if !c.skipNodeAdd {
		maker := bytes.NewBuffer(nil)
		for i := 0; i < c.nodesCount; i++ {
			if i > 0 {
				maker.WriteRune(',')
			}
			fmt.Fprintf(maker, "%sn%d", common.GetObjectPrefix(), i+1)
		}
		code, err = manager.Exec(ctx, node1ID, "shardmanctl nodes add -n "+maker.String(), "postgres")
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("command status code: %d", code)
		}
	}

	log.Printf("mount: %s --> /mntdata", common.GetVolumeDir())

	return nil
}

func (c *CommandStart) isStarted(ctx context.Context, manager *container.Manager, networker *network.Manager) error {
	if c.force {
		sc, err := stop.NewCase(manager, networker)
		if err != nil {
			return err
		}
		if err = sc.Exec(ctx); err != nil {
			return err
		}
	} else {
		return errors.New("already started")
	}

	return nil
}

func (c *CommandStart) update(ctx context.Context) error {
	builder, err := image.NewManager(c.cli)
	if err != nil {
		return fmt.Errorf("image build init error: %w", err)
	}

	myCase, err := build.NewCase(builder)
	if err != nil {
		return fmt.Errorf("rebuild init case error: %w", err)
	}

	return myCase.Exec(ctx)
}
