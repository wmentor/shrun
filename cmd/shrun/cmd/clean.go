package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/clean"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandClean)(nil)
)

type CommandClean struct {
	command *cobra.Command
	cli     *client.Client
	all     bool
}

func NewCommandClean(cli *client.Client) *CommandClean {
	cp := &CommandClean{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "clean",
		Short: "delete images",
		RunE:  cp.exec,
	}

	cc.Flags().BoolVarP(&cp.all, "all", "a", false, "delete base images")

	cp.command = cc

	return cp
}

func (cp *CommandClean) Command() *cobra.Command {
	return cp.command
}

func (cp *CommandClean) exec(cc *cobra.Command, _ []string) error {
	imageManager, err := image.NewManager(cp.cli)
	if err != nil {
		return err
	}

	myCase, err := clean.NewCase(imageManager)
	if err != nil {
		return err
	}

	return myCase.Exec(cc.Context())
}
