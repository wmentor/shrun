package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/cmd/shrun/cmd/stolon"
)

var (
	_ cmd.CobraCommand = (*CommandPull)(nil)
)

type CommandStolon struct {
	command *cobra.Command
	cli     *client.Client
}

func NewCommandStolon(cli *client.Client) *CommandStolon {
	ci := &CommandStolon{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "stolon",
		Short: "stolon [subcommand]",
	}

	cc.AddCommand(stolon.NewCommandITest(cli).Command())

	ci.command = cc

	return ci
}

func (ci *CommandStolon) Command() *cobra.Command {
	return ci.command
}
