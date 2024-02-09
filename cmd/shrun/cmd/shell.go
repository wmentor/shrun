package cmd

import (
	"errors"
	"fmt"
	"strings"

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
	command  *cobra.Command
	cli      *client.Client
	node     string
	user     string
	debugCmd string
	execCmd  string
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
	cc.Flags().StringVarP(&ci.user, "user", "u", common.PgUser, "user name")
	cc.Flags().StringVarP(&ci.execCmd, "cmd", "c", "", "exec command")
	cc.Flags().StringVarP(&ci.debugCmd, "debug", "d", "", "debug command")

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

	cmd := common.CmdBash
	if ci.user == common.PgUser && ci.execCmd == "" && ci.debugCmd == "" {
		cmd = append(cmd, "-c", "/var/lib/postgresql/.build_info && /bin/bash")
	}

	if ci.execCmd != "" {
		cmd = append([]string{}, cmd...)
		cmd = append(cmd, "-c", ci.execCmd)
	}

	if ci.debugCmd != "" && ci.execCmd == "" {
		debugCMD := strings.Split(ci.debugCmd, " ")
		args := debugCMD[1:]

		for i := range args {
			if strings.HasPrefix(args[i], "-") {
				args = append(args[:i], append([]string{"--"}, args[i:]...)...)
				break
			}
		}

		debugCommand := strings.Join(append([]string{fmt.Sprintf(
			`dlv --listen=:40000 --headless=true --api-version=2 exec $(which %s)`, debugCMD[0])},
			args...), " ")

		cmd = append(cmd, "-c", debugCommand)
	}

	return mng.ShellCommand(cc.Context(), ci.node, ci.user, "", cmd)
}
