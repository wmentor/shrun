package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/pull"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandPull)(nil)
)

type CommandPull struct {
	command *cobra.Command
	cli     *client.Client
	force   bool
}

func NewCommandPull(cli *client.Client) *CommandPull {
	cp := &CommandPull{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "pull",
		Short: "pull images (golang:1.18, ubuntu:20.04)",
		RunE:  cp.exec,
	}

	cc.Flags().BoolVarP(&cp.force, "force", "f", false, "force pull images if already exists")

	cp.command = cc

	return cp
}

func (cp *CommandPull) Command() *cobra.Command {
	return cp.command
}

func (cp *CommandPull) exec(cc *cobra.Command, _ []string) error {
	imageManager, err := image.NewManager(cp.cli)
	if err != nil {
		return err
	}

	myCase, err := pull.NewCase(imageManager, cp.force)
	if err != nil {
		return err
	}

	return myCase.Exec(cc.Context())
}
