package cmd

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/gobuilder"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandGoBuilder)(nil)
)

type CommandGoBuilder struct {
	command *cobra.Command
	cli     *client.Client
	rebuild bool
}

func NewCommandGoBuilder(cli *client.Client) *CommandGoBuilder {
	ci := &CommandGoBuilder{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "gobuilder",
		Short: "Run gobuilder",
		RunE:  ci.exec,
	}

	cc.Flags().BoolVarP(&ci.rebuild, "rebuild", "r", false, "rebuilder gobuilder")

	ci.command = cc

	return ci
}

func (ci *CommandGoBuilder) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandGoBuilder) exec(cc *cobra.Command, _ []string) error {
	mng, err := container.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	img, err := image.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create image builder error: %w", err)
	}

	myCase, err := gobuilder.NewCase(img, mng)
	if err != nil {
		return err
	}

	if ci.rebuild {
		myCase.WithImageRebuild()
	}

	return myCase.Exec(cc.Context())
}
