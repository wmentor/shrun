package cmd

import (
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandBuild)(nil)
)

type CommandBuild struct {
	command     *cobra.Command
	cli         *client.Client
	noGoProxy   bool
	repfactor   int
	topology    string
	clusterName string
	etcdCount   int
	logLevel    string
	pgMajor     int
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

	ctx := cc.Context()

	for _, copyFile := range []string{image.SpecFile, image.RcLocal} {
		if err = common.CopyFile(ctx, filepath.Join(common.GetConfigDir(), copyFile), filepath.Join(common.GetDataDir(), copyFile)); err != nil {
			return err
		}
	}

	if err = imageManager.BuildImage(ctx, image.DockerfileGoBuilder, "gobuilder:latest"); err != nil {
		return err
	}

	if err = imageManager.BuildImage(ctx, image.DockerfilePgBuildEnv, "pgbuildenv:latest"); err != nil {
		return err
	}

	if err = imageManager.BuildImage(ctx, image.DockerfilePgDestEnv, "pgdestenv:latest"); err != nil {
		return err
	}

	if err = imageManager.BuildImage(ctx, image.DockerFileSdmNode, "sdmnode:latest"); err != nil {
		return err
	}

	if err = imageManager.BuildImage(ctx, image.DockerFileShardman, "shardman:latest"); err != nil {
		return err
	}

	return nil
}
