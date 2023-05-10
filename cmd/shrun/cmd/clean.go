package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/clean"
	"github.com/wmentor/shrun/internal/cases/stop"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/image"
	"github.com/wmentor/shrun/internal/network"
)

var (
	_ cmd.CobraCommand = (*CommandClean)(nil)
)

type CommandClean struct {
	command *cobra.Command
	cli     *client.Client
	force   bool
}

func NewCommandClean(cli *client.Client) *CommandClean {
	cp := &CommandClean{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "clean",
		Short: "delete images",
		RunE:  cp.exec,
	}

	cc.Flags().BoolVarP(&cp.force, "force", "f", false, "stopping cluster")

	cp.command = cc

	return cp
}

func (c *CommandClean) Command() *cobra.Command {
	return c.command
}

func (c *CommandClean) exec(cc *cobra.Command, _ []string) error {
	if c.force {
		manager, err := container.NewManager(c.cli)
		if err != nil {
			return err
		}

		networker, err := network.NewManager(c.cli)
		if err != nil {
			return err
		}

		stopCase, err := stop.NewCase(manager, networker)
		if err != nil {
			return err
		}
		if err := stopCase.Exec(cc.Context()); err != nil {
			return err
		}
	}
	imageManager, err := image.NewManager(c.cli)
	if err != nil {
		return err
	}
	myCase, err := clean.NewCase(imageManager)
	if err != nil {
		return err
	}

	return myCase.WithForce(c.force).Exec(cc.Context())
}
