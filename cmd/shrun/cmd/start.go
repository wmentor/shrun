package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/network"
)

var (
	_ cmd.CobraCommand = (*CommandBuild)(nil)
)

type CommandStart struct {
	command     *cobra.Command
	cli         *client.Client
	nodesCount  int
	makeCluster bool
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
	cc.Flags().BoolVar(&c.makeCluster, "make-cluster", false, "make cluster")

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

	started, err := networker.CheckNetworkExists(ctx)
	if err != nil {
		return err
	}

	if started {
		return errors.New("already started")
	}

	netID, err := networker.CreateNetwork(ctx)
	if err != nil {
		return fmt.Errorf("create network error: %w", err)
	}

	manager, err := container.NewManager(c.cli)
	if err != nil {
		return err
	}

	etcdList, err := common.GetEtcdList()
	if err != nil {
		return err
	}

	etcdClusterMaker := strings.Builder{}
	for i := range etcdList {
		hostname := fmt.Sprintf("%se%d", common.GetObjectPrefix(), i+1)
		if i == 0 {
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
			"/opt/pgpro/etcd/bin/etcd",
			fmt.Sprintf("--name=%s", hostname),
			fmt.Sprintf("--initial-advertise-peer-urls=http://%s:2380", hostname),
			"--listen-peer-urls=http://0.0.0.0:2380",
			fmt.Sprintf("--advertise-client-urls=http://%s:2379", hostname),
			"--listen-client-urls=http://0.0.0.0:2379",
			fmt.Sprintf("--initial-cluster=%s", etcdClusterMaker.String()),
		}

		log.Printf("start %s", hostname)

		opts := container.ContainerStartSettings{
			Image:     "etcd:latest",
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

	envs := []string{
		fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
		fmt.Sprintf("SDM_CLUSTER_NAME=%s", clusterName),
		"SDM_LOG_LEVEL=debug",
		fmt.Sprintf("SDM_STORE_ENDPOINTS=%s", strings.Join(etcdList, ",")),
	}

	for i := 0; i < c.nodesCount; i++ {
		hostname := fmt.Sprintf("%sn%d", common.GetObjectPrefix(), i+1)
		log.Printf("start %s", hostname)

		opts := container.ContainerStartSettings{
			Image:     "shardman:latest",
			Host:      hostname,
			NetworkID: netID,
			Envs:      envs,
		}

		if id, err := manager.CreateAndStart(ctx, opts); err == nil {
			containerIDs[hostname] = id
		} else {
			log.Println("failed")
		}
	}

	log.Println("load sdmspec.json")
	manager.Exec(ctx, containerIDs["shrn1"], []string{"sh", "-c", "shardmanctl init -f /etc/shardman/sdmspec.json"})

	if c.makeCluster {
		maker := bytes.NewBuffer(nil)
		for i := 0; i < c.nodesCount; i++ {
			if i > 0 {
				maker.WriteRune(',')
			}
			fmt.Fprintf(maker, "%sn%d", common.GetObjectPrefix(), i+1)
		}
		log.Println("add nodes")
		manager.Exec(ctx, containerIDs["shrn1"], []string{"sh", "-c", "shardmanctl nodes add -n " + maker.String()})
	}

	return nil
}
