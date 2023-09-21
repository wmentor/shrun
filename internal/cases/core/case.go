package core

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

func (c *Case) Exec(ctx context.Context) error {
	contName := "core"
	imgName := fmt.Sprintf("%s:latest", contName)

	if err := c.contManager.RemoveContainer(ctx, contName); err != nil {
		return err
	}

	opts := entities.ContainerStartSettings{
		Image: imgName,
		Host:  contName,
	}

	cid, err := c.contManager.CreateAndStart(ctx, opts)
	if err != nil {
		return fmt.Errorf("create and start %s error: %w", contName, err)
	}

	c.contManager.Exec(ctx, cid, "cd /repo/shardman && make clean || true", common.PgUser)

	configure := "cd /repo/shardman && ./configure --enable-debug --enable-cassert --enable-nls --with-perl " +
		"--with-python --with-tcl --with-openssl --with-libxml --with-libxslt --with-ldap --with-icu --with-tclconfig=/usr/lib/tcl8.6 --enable-svt5"

	c.contManager.Exec(ctx, cid, configure, common.PgUser)

	c.contManager.ShellCommand(ctx, contName, "postgres", "/repo/shardman", common.CmdBash)

	if err := c.contManager.RemoveContainer(ctx, contName); err != nil {
		return err
	}

	return nil
}
