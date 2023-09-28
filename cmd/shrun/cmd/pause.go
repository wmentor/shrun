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
	_ cmd.CobraCommand = (*CommandPause)(nil)
)

type CommandPause struct {
	command *cobra.Command
	cli     *client.Client
}

func NewCommandPause(cli *client.Client) *CommandPause {
	c := &CommandPause{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "pause",
		Short: "pause all containers",
		RunE:  c.exec,
	}

	c.command = cc

	return c
}

func (c *CommandPause) Command() *cobra.Command {
	return c.command
}

func (c *CommandPause) exec(cc *cobra.Command, _ []string) error {
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

	return myCase.WithRemove(false).Exec(cc.Context())
}
