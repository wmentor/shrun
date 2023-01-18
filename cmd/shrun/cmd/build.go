package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/build"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandBuild)(nil)
)

type CommandBuild struct {
	command       *cobra.Command
	cli           *client.Client
	buildBasic    bool
	buildPostgres bool
}

func NewCommandBuild(cli *client.Client) *CommandBuild {
	cb := &CommandBuild{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "build",
		Short: "build Dockerfiles",
		RunE:  cb.exec,
	}

	cc.Flags().BoolVar(&cb.buildBasic, "build-basic", false, "build gobuild, pgbuildenv, pgdestenv")
	cc.Flags().BoolVar(&cb.buildPostgres, "build-pg", false, "build postgres")

	cb.command = cc

	return cb
}

func (cb *CommandBuild) Command() *cobra.Command {
	return cb.command
}

func (cb *CommandBuild) exec(cc *cobra.Command, _ []string) error {
	imageManager, err := image.NewManager(cb.cli)
	if err != nil {
		return err
	}

	myCase, err := build.NewCase(imageManager)
	if err != nil {
		return err
	}

	if cb.buildBasic {
		myCase.WithBuildBasic()
	}

	if cb.buildPostgres {
		myCase.WithBuildPG()
	}

	return myCase.Exec(cc.Context())
}
