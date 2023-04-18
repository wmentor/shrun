package cmd

import (
	"errors"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/container"
)

var (
	_ cmd.CobraCommand = (*CommandPull)(nil)
)

type CommandShell struct {
	command             *cobra.Command
	cli                 *client.Client
	node                string
	user                string
	shardmanctlDebugCmd string
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

	cc.Flags().StringVarP(&ci.node, "node", "n", "", "node name")
	cc.Flags().StringVarP(&ci.user, "user", "u", "postgres", "user name")
	cc.Flags().StringVar(&ci.shardmanctlDebugCmd, "shardmanctl-debug", "", "debug shardmanctl command:"+
		"such as status")

	ci.command = cc

	return ci
}

func (ci *CommandShell) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandShell) exec(cc *cobra.Command, _ []string) error {
	if ci.node == "" {
		ci.node = common.GetNodeName(1)
	}

	if ci.user == "" {
		return errors.New("unknown username")
	}

	mng, err := container.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	cmd := []string{"/bin/bash"}

	if ci.shardmanctlDebugCmd != "" {
		debugCommand := "dlv --listen=:40000 --headless=true --api-version=2 exec $(which shardmanctl)"

		cmd = append(cmd, "-c", fmt.Sprintf(`%s %s`, debugCommand, ci.shardmanctlDebugCmd))
	}

	return mng.ShellCommand(cc.Context(), ci.node, ci.user, cmd)
}
