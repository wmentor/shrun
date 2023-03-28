package cmd

import (
	"errors"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/nodes/psql"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/container"
)

var (
	_ cmd.CobraCommand = (*CommandPSQL)(nil)
)

type CommandPSQL struct {
	command *cobra.Command
	cli     *client.Client
	node    string
	port    int
}

func NewCommandPSQL(cli *client.Client) *CommandPSQL {
	ci := &CommandPSQL{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "psql",
		Short: "Run psql to connect",
		RunE:  ci.exec,
	}

	cc.Flags().StringVarP(&ci.node, "node", "n", "", "node hostname")
	cc.Flags().IntVarP(&ci.port, "port", "p", 0, "database port (default from sdmspec.json)")

	ci.command = cc

	return ci
}

func (ci *CommandPSQL) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandPSQL) exec(cc *cobra.Command, _ []string) error {
	if ci.node == "" {
		ci.node = common.GetNodeName(1)
	}

	if ci.port < 0 || ci.port > 0xffff {
		return errors.New("invalid port")
	}

	mng, err := container.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	myCase, err := psql.NewCase(mng, ci.node)
	if err != nil {
		return err
	}

	if ci.port != 0 {
		myCase.WithPort(ci.port)
	}

	return myCase.Exec(cc.Context())
}
