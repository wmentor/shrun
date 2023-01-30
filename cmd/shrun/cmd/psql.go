package cmd

import (
	"errors"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/nodes/psql"
	"github.com/wmentor/shrun/internal/container"
)

var (
	_ cmd.CobraCommand = (*CommandPSQL)(nil)
)

type CommandPSQL struct {
	command *cobra.Command
	cli     *client.Client
	node    string
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

	cc.Flags().StringVarP(&ci.node, "node", "n", "shrn1", "node hostname")

	ci.command = cc

	return ci
}

func (ci *CommandPSQL) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandPSQL) exec(cc *cobra.Command, _ []string) error {
	if ci.node == "" {
		return errors.New("invalid node")
	}

	mng, err := container.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	myCase, err := psql.NewCase(mng, ci.node)
	if err != nil {
		return err
	}

	return myCase.Exec(cc.Context())
}
