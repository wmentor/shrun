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
	command       *cobra.Command
	cli           *client.Client
	buildPgDoc    bool
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

	cc.Flags().BoolVar(&cb.buildPgDoc, "build-pg-doc", false, "build pgdoc")
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

	ctx := cc.Context()

	for _, copyFile := range []string{image.SpecFile, image.RcLocal} {
		if err = common.CopyFile(ctx, filepath.Join(common.GetConfigDir(), copyFile), filepath.Join(common.GetDataDir(), copyFile)); err != nil {
			return err
		}
	}

	if cb.buildBasic {
		if err = imageManager.BuildImage(ctx, image.DockerfileGoBuilder, "gobuilder:latest"); err != nil {
			return err
		}

		if err = imageManager.BuildImage(ctx, image.DockerfilePgBuildEnv, "pgbuildenv:latest"); err != nil {
			return err
		}

		if err = imageManager.BuildImage(ctx, image.DockerfilePgDestEnv, "pgdestenv:latest"); err != nil {
			return err
		}

		if err = imageManager.BuildImage(ctx, image.DockerfileEtcd, "etcd:latest"); err != nil {
			return err
		}
	}

	if cb.buildPostgres {
		if err = imageManager.BuildImage(ctx, image.DockerfileSdmNode, "sdmnode:latest"); err != nil {
			return err
		}
	}

	if err = imageManager.BuildImage(ctx, image.DockerfileShardman, "shardman:latest"); err != nil {
		return err
	}

	if cb.buildPgDoc {
		if err = imageManager.BuildImage(ctx, image.DockerfilePgDoc, "pgdoc:latest"); err != nil {
			return err
		}
	}

	return nil
}
