package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/cmd/shrun/cmd/nodes"
)

var (
	_ cmd.CobraCommand = (*CommandPull)(nil)
)

type CommandNodes struct {
	command *cobra.Command
	cli     *client.Client
}

func NewCommandNodes(cli *client.Client) *CommandNodes {
	ci := &CommandNodes{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "nodes",
		Short: "nodes [subcommand]",
	}

	cc.AddCommand(nodes.NewCommandAdd(cli).Command())
	cc.AddCommand(nodes.NewCommandRM(cli).Command())

	ci.command = cc

	return ci
}

func (ci *CommandNodes) Command() *cobra.Command {
	return ci.command
}
