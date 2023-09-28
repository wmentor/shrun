package cmd

import (
	"errors"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/network"
)

var (
	_ cmd.CobraCommand = (*CommandRenew)(nil)
)

type CommandRenew struct {
	command *cobra.Command
	cli     *client.Client
}

func NewCommandRenew(cli *client.Client) *CommandRenew {
	c := &CommandRenew{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "renew",
		Short: "renew all containers",
		RunE:  c.exec,
	}

	c.command = cc

	return c
}

func (c *CommandRenew) Command() *cobra.Command {
	return c.command
}

func (c *CommandRenew) exec(cc *cobra.Command, _ []string) error {
	manager, err := container.NewManager(c.cli)
	if err != nil {
		return err
	}

	networker, err := network.NewManager(c.cli)
	if err != nil {
		return err
	}

	ctx := cc.Context()

	started, err := networker.CheckNetworkExists(ctx)
	if err != nil {
		return err
	}

	if !started {
		return errors.New("containers not found")
	}

	return manager.Renew(ctx)
}
