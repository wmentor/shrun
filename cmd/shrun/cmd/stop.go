package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/stop"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/network"
)

var (
	_ cmd.CobraCommand = (*CommandStop)(nil)
)

type CommandStop struct {
	command *cobra.Command
	cli     *client.Client
}

func NewCommandStop(cli *client.Client) *CommandStop {
	c := &CommandStop{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "stop",
		Short: "stop and remove all entities",
		RunE:  c.exec,
	}

	c.command = cc

	return c
}

func (c *CommandStop) Command() *cobra.Command {
	return c.command
}

func (c *CommandStop) exec(cc *cobra.Command, _ []string) error {
	manager, err := container.NewManager(c.cli)
	if err != nil {
		return err
	}

	networker, err := network.NewManager(c.cli)
	if err != nil {
		return err
	}

	myCase, err := stop.NewCase(manager, networker)
	if err != nil {
		return err
	}

	return myCase.Exec(cc.Context())
}
