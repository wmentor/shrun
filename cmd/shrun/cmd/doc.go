package cmd

import (
	"log"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/common"
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

	ctx := cc.Context()

	if err = containerManager.RemoveContainer(ctx, "pgdoc"); err != nil {
		return err
	}

	if err = imageManager.BuildImage(ctx, image.DockerfilePgDoc, "pgdoc:latest"); err != nil {
		return err
	}

	opts := container.ContainerStartSettings{
		Image: "pgdoc:latest",
		Host:  "pgdoc",
	}

	cid, err := containerManager.CreateAndStart(ctx, opts)
	if err != nil {
		log.Printf("start container error: %v", err)
		return err
	}

	containerManager.Exec(ctx, cid, "mkdir -p /mntdata/doc ; rm -rf /mntdata/doc/* ; cp -r /build/shardman/contrib/shardman/doc/html/ /mntdata/doc", "root")

	if err = containerManager.RemoveContainer(ctx, "pgdoc"); err != nil {
		return err
	}

	log.Printf("doc was saved to %s/doc/html", common.GetVolumeDir())

	return nil
}
