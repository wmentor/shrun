package gobuilder

import (
	"context"
	"errors"
	"fmt"

	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/entities"
)

var (
	ErrInvalidImageBuidler     = errors.New("invalid image builder")
	ErrInvalidContainerManager = errors.New("invalid container manager")
)

type Case struct {
	imgBuilder  ImageBuilder
	contManager ContainerManager
	rebuilder   bool
}

func NewCase(imgBuilder ImageBuilder, contManager ContainerManager) (*Case, error) {
	if imgBuilder == nil || imgBuilder == ImageBuilder(nil) {
		return nil, ErrInvalidImageBuidler
	}

	if contManager == nil || contManager == ContainerManager(nil) {
		return nil, ErrInvalidContainerManager
	}

	c := &Case{
		imgBuilder:  imgBuilder,
		contManager: contManager,
	}

	return c, nil
}

func (c *Case) WithImageRebuild() *Case {
	c.rebuilder = true
	return c
}

func (c *Case) Exec(ctx context.Context) error {
	contName := "gobuilder"
	imgName := fmt.Sprintf("%s:latest", contName)

	if err := c.contManager.RemoveContainer(ctx, contName); err != nil {
		return err
	}

	if c.rebuilder {
		if err := c.imgBuilder.BuildImage(ctx, common.DockerfileGoBuilder, imgName); err != nil {
			return err
		}
	}

	opts := entities.ContainerStartSettings{
		Image: imgName,
		Host:  contName,
	}

	if _, err := c.contManager.CreateAndStart(ctx, opts); err != nil {
		return fmt.Errorf("create and start %s error: %w", contName, err)
	}

	c.contManager.ShellCommand(ctx, contName, "root", "/repo", common.CmdBash)

	if err := c.contManager.RemoveContainer(ctx, contName); err != nil {
		return err
	}

	return nil
}
