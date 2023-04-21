package cmd

import (
	"errors"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/gotpc"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/image"
	"github.com/wmentor/shrun/internal/network"
)

var (
	_ cmd.CobraCommand = (*CommandGoTpc)(nil)
)

type CommandGoTpc struct {
	command *cobra.Command
	cli     *client.Client
	rebuild bool
}

func NewCommandGoTpc(cli *client.Client) *CommandGoTpc {
	ci := &CommandGoTpc{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "gotpc",
		Short: "Run gobuilder",
		RunE:  ci.exec,
	}

	cc.Flags().BoolVarP(&ci.rebuild, "rebuild", "r", false, "rebuilder gobuilder")

	ci.command = cc

	return ci
}

func (ci *CommandGoTpc) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandGoTpc) exec(cc *cobra.Command, _ []string) error {
	mng, err := container.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	img, err := image.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create image builder error: %w", err)
	}

	networker, err := network.NewManager(ci.cli)
	if err != nil {
		return err
	}

	ctx := cc.Context()

	start, err := networker.CheckNetworkExists(ctx)
	if err != nil {
		return err
	}

	if !start {
		return errors.New("network not started or cluster not found.")
	}

	netID, err := networker.GetNetworkID(ctx)
	if err != nil {
		return err
	}

	myCase, err := gotpc.NewCase(img, mng, netID)
	if err != nil {
		return err
	}

	if ci.rebuild {
		myCase.WithImageRebuild()
	}

	return myCase.Exec(ctx)
}
