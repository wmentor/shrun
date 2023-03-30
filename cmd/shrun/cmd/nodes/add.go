package nodes

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/entities"
	"github.com/wmentor/shrun/internal/network"
)

var (
	_ cmd.CobraCommand = (*CommandAdd)(nil)
)

type CommandAdd struct {
	command     *cobra.Command
	cli         *client.Client
	nodesCount  int
	memoryLimit string
	cpuLimit    float64
	mountData   bool
}

func NewCommandAdd(cli *client.Client) *CommandAdd {
	ci := &CommandAdd{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "add",
		Short: "Add nodes command",
		RunE:  ci.exec,
	}

	cc.Flags().StringVar(&ci.memoryLimit, "memory", "", "memory limit")
	cc.Flags().Float64Var(&ci.cpuLimit, "cpu", 0, "cpu limit")
	cc.Flags().IntVarP(&ci.nodesCount, "nodes", "n", 1, "add nodes count")
	cc.Flags().BoolVar(&ci.mountData, "mount-data", false, "mount pg data to builddir/pgdata/hostname")

	ci.command = cc

	return ci
}

func (ci *CommandAdd) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandAdd) exec(cc *cobra.Command, _ []string) error {
	ctx := cc.Context()

	if ci.nodesCount < 1 {
		return errors.New("invalid node count")
	}

	mng, err := container.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	nextNum := 0

	for {
		nextNum++

		nodeName := common.GetNodeName(nextNum)

		_, err := mng.GetContainer(ctx, nodeName)
		if err != nil {
			if errors.Is(err, common.ErrNotFound) {
				break
			}
			return err
		}
	}

	networker, err := network.NewManager(ci.cli)
	if err != nil {
		return err
	}

	start, err := networker.CheckNetworkExists(ctx)
	if err != nil {
		return err
	}

	if !start {
		return errors.New("network not started (maybe you need start command).")
	}

	netID, err := networker.GetNetworkID(ctx)
	if err != nil {
		return err
	}

	etcdList, err := common.GetEtcdList()
	if err != nil {
		return err
	}

	clusterName, _ := common.GetClusterName()
	logLevel, _ := common.GetLogLevel()

	envs := []string{
		fmt.Sprintf("CLUSTER_NAME=%s", clusterName),
		fmt.Sprintf("SDM_CLUSTER_NAME=%s", clusterName),
		fmt.Sprintf("SDM_LOG_LEVEL=%s", logLevel),
		fmt.Sprintf("SDM_STORE_ENDPOINTS=%s", strings.Join(etcdList, ",")),
	}

	for i := 0; i < ci.nodesCount; i++ {
		hostname := common.GetNodeName(nextNum + i)
		log.Printf("start %s", hostname)

		opts := entities.ContainerStartSettings{
			Image:     "shardman:latest",
			Host:      hostname,
			NetworkID: netID,
			Envs:      envs,
			MountData: ci.mountData,
		}

		if ci.memoryLimit != "" {
			size, err := units.RAMInBytes(ci.memoryLimit)
			if err != nil {
				return fmt.Errorf("invalid memory limit: %v", ci.memoryLimit)
			}
			opts.MemoryLimit = size
		}

		if ci.cpuLimit != 0 {
			opts.CPU = ci.cpuLimit
		}

		if _, err := mng.CreateAndStart(ctx, opts); err != nil {
			return fmt.Errorf("create and start %s error: %w", hostname, err)
		}
	}

	return nil
}
