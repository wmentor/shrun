package cmd

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/build"
	"github.com/wmentor/shrun/internal/cases/core"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandCore)(nil)
)

type CommandCore struct {
	command *cobra.Command
	cli     *client.Client
}

func NewCommandCore(cli *client.Client) *CommandCore {
	cb := &CommandCore{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "core",
		Short: "shardman core build environment",
		RunE:  cb.exec,
	}

	cb.command = cc

	return cb
}

func (cb *CommandCore) Command() *cobra.Command {
	return cb.command
}

func (cb *CommandCore) exec(cc *cobra.Command, _ []string) error {
	imageManager, err := image.NewManager(cb.cli)
	if err != nil {
		return err
	}

	bCase, err := build.NewCase(imageManager)
	if err != nil {
		return err
	}

	if err = bCase.WithBuildBasic().WithCore().Exec(cc.Context()); err != nil {
		return err
	}

	mng, err := container.NewManager(cb.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	rCase, err := core.NewCase(imageManager, mng)
	if err != nil {
		return err
	}

	return rCase.Exec(cc.Context())
}
