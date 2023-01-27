package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/doc"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandBuild)(nil)
)

type CommandDoc struct {
	command *cobra.Command
	cli     *client.Client
}

func NewCommandDoc(cli *client.Client) *CommandDoc {
	cb := &CommandDoc{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "doc",
		Short: "build shardman html doc",
		RunE:  cb.exec,
	}

	cb.command = cc

	return cb
}

func (cb *CommandDoc) Command() *cobra.Command {
	return cb.command
}

func (cb *CommandDoc) exec(cc *cobra.Command, _ []string) error {
	imageManager, err := image.NewManager(cb.cli)
	if err != nil {
		return err
	}

	containerManager, err := container.NewManager(cb.cli)
	if err != nil {
		return err
	}

	myCase, err := doc.NewCase(imageManager, containerManager)
	if err != nil {
		return err
	}

	return myCase.Exec(cc.Context())
}
