package clean

import (
	"context"
	"errors"

	"github.com/wmentor/shrun/internal/common"
)

var (
	ErrInvalidManager = errors.New("invalid image manager")
)
var (
	imageList = []string{common.GetNodeContainerName(), common.GetSdmNodeImageName(),
		"gobuilder:latest", "pgbuildenv:latest", "pgdestenv:latest", "gotpc:latest", "pgdoc:latest"}
)

var (
	requiredImages    = []string{"postgres:14", "ubuntu:20.04", "golang:1.20.5", "quay.io/coreos/etcd:v3.5.8"}
	armRequiredImages = []string{"arm64v8/postgres:14", "arm64v8/ubuntu:20.04", "golang:1.20.5", "quay.io/coreos/etcd:v3.5.8-arm64"}
)

type Case struct {
	mng   Cleaner
	force bool
}

func NewCase(mng Cleaner) (*Case, error) {
	if mng == nil || mng == Cleaner(nil) {
		return nil, ErrInvalidManager
	}
	return &Case{
		mng: mng,
	}, nil
}

func (c *Case) WithForce(force bool) *Case {
	c.force = force
	return c
}

func (c *Case) Exec(ctx context.Context) error {
	dockerImages := imageList
	if common.WorkArch == common.ArchArm64 {
		dockerImages = append(dockerImages, armRequiredImages...)
	} else {
		dockerImages = append(dockerImages, requiredImages...)

	}
	for _, image := range dockerImages {
		if err := c.mng.RemoveImage(ctx, image, c.force); err != nil {
			return err
		}
	}
	if err := c.mng.BuilderPrune(ctx, true); err != nil {
		return err
	}
	return nil
}
