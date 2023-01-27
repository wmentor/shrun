package doc

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/entities"
)

var (
	ErrInvalidContainerManager = errors.New("invalid container manager")
	ErrInvalidContainerBuilder = errors.New("invalid container builder")
)

const (
	containerName = "pgdoc"
	imageName     = "pgdoc:latest"
	docDir        = "/mntdata/doc"
)

type Case struct {
	builder Builder
	manager ContainerManager
}

func NewCase(builder Builder, manager ContainerManager) (*Case, error) {
	if builder == nil || builder == Builder(nil) {
		return nil, ErrInvalidContainerBuilder
	}

	if manager == nil || manager == ContainerManager(nil) {
		return nil, ErrInvalidContainerManager
	}

	return &Case{
		builder: builder,
		manager: manager,
	}, nil
}

func (c *Case) Exec(ctx context.Context) error {
	if err := c.manager.RemoveContainer(ctx, containerName); err != nil {
		return err
	}

	if err := c.builder.BuildImage(ctx, common.DockerfilePgDoc, imageName); err != nil {
		return err
	}

	opts := entities.ContainerStartSettings{
		Image: imageName,
		Host:  containerName,
	}

	cid, err := c.manager.CreateAndStart(ctx, opts)
	if err != nil {
		log.Printf("start container error: %v", err)
		return err
	}

	cmd := fmt.Sprintf("mkdir -p %s ; rm -rf %s/* ; cp -r /build/shardman/contrib/shardman/doc/html/ %s", docDir, docDir, docDir)

	c.manager.Exec(ctx, cid, cmd, "root")

	if err = c.manager.RemoveContainer(ctx, containerName); err != nil {
		return err
	}

	log.Printf("doc was saved to %s/doc/html", common.GetVolumeDir())

	return nil
}
