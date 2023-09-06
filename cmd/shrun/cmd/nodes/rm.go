package nodes

import (
	"errors"
	"fmt"
	"log"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/container"
)

var (
	_ cmd.CobraCommand = (*CommandAdd)(nil)
)

type CommandRM struct {
	command    *cobra.Command
	cli        *client.Client
	nodesCount int
}

func NewCommandRM(cli *client.Client) *CommandRM {
	ci := &CommandRM{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "rm",
		Short: "Remove nodes command",
		RunE:  ci.exec,
	}

	cc.Flags().IntVarP(&ci.nodesCount, "nodes", "n", 1, "remove nodes count")

	ci.command = cc

	return ci
}

func (ci *CommandRM) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandRM) exec(cc *cobra.Command, _ []string) error {
	ctx := cc.Context()

	if ci.nodesCount < 1 {
		return errors.New("invalid node count")
	}

	mng, err := container.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	nextNum := 0
	nodes := make([]string, 0, 16)
	gnodes := make([]string, 0, 16)

	hasGrafana := common.GetGrafanaStatus()

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
		nodes = append(nodes, nodeName)

		if hasGrafana {
			gnodes = append(gnodes, mng.GetExporterName(nextNum))
		}
	}

	if len(nodes) == 0 {
		log.Println("nodes not found")
	}

	if ci.nodesCount < len(nodes) {
		nodes = nodes[len(nodes)-ci.nodesCount:]
	}

	if ci.nodesCount < len(gnodes) {
		gnodes = gnodes[len(gnodes)-ci.nodesCount:]
	}

	err = nil

	for _, node := range nodes {
		if curErr := mng.RemoveContainer(ctx, node); curErr != nil {
			if err == nil {
				curErr = err
			}
		}
	}

	for _, node := range gnodes {
		mng.RemoveContainer(ctx, node)
	}

	return err
}
