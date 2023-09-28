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
	_ cmd.CobraCommand = (*CommandResume)(nil)
)

type CommandResume struct {
	command *cobra.Command
	cli     *client.Client
}

func NewCommandResume(cli *client.Client) *CommandResume {
	c := &CommandResume{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "resume",
		Short: "resume all containers",
		RunE:  c.exec,
	}

	c.command = cc

	return c
}

func (c *CommandResume) Command() *cobra.Command {
	return c.command
}

func (c *CommandResume) exec(cc *cobra.Command, _ []string) error {
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
