package cmd

import (
	"errors"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/container"
)

var (
	_ cmd.CobraCommand = (*CommandPull)(nil)
)

type CommandShell struct {
	command *cobra.Command
	cli     *client.Client
	node    string
	user    string
}

func NewCommandShell(cli *client.Client) *CommandShell {
	ci := &CommandShell{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "shell",
		Short: "open node shell",
		RunE:  ci.exec,
	}

	cc.Flags().StringVarP(&ci.node, "node", "n", "shrn1", "node name")
	cc.Flags().StringVarP(&ci.user, "user", "u", "root", "user name")

	ci.command = cc

	return ci
}

func (ci *CommandShell) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandShell) exec(cc *cobra.Command, _ []string) error {
	if ci.node == "" {
		return errors.New("unknown node name")
	}

	if ci.user == "" {
		return errors.New("unknown username")
	}

	return container.Shell(cc.Context(), ci.node, ci.user)
}
