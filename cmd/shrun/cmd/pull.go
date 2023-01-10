package cmd

import (
	"errors"
	"log"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/common"
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
		Short: "pull images (golang:1.18, etcd:latest, ubuntu:20.04, bitnami/etcd:3.5.6)",
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

	ctx := cc.Context()

	imgNames := []string{"bitnami/etcd:3.5.6", "postgres:14", "ubuntu:20.04", "golang:1.18"}

	for _, img := range imgNames {
		if !cp.force {
			if err = imageManager.CheckImageExists(ctx, img); err == nil {
				log.Printf("image %s found. skip\n", img)
				continue
			}

			if !errors.Is(err, common.ErrNotFound) {
				return err
			}
		}

		log.Printf("pull image %s\n", img)
		if err = imageManager.PullImage(ctx, img); err != nil {
			log.Println(err)
		}
	}

	return nil
}
