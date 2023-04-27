package pull

import (
	"context"
	"errors"
	"log"

	"github.com/wmentor/shrun/internal/common"
)

var (
	ErrInvalidImageManager = errors.New("invalid image manager")
)

var (
	requiredImages = []string{"postgres:14", "ubuntu:20.04", "golang:1.18", "quay.io/coreos/etcd:v3.5.8"}
)

var (
	armRequiredImages = []string{"arm64v8/postgres:14", "arm64v8/ubuntu:20.04", "golang:1.18", "quay.io/coreos/etcd:v3.5.8-arm64"}
)

type Case struct {
	mng       ImageManager
	forcePull bool
}

func NewCase(mng ImageManager, forcePull bool) (*Case, error) {
	if mng == nil || mng == ImageManager(nil) {
		return nil, ErrInvalidImageManager
	}

	c := &Case{
		mng:       mng,
		forcePull: forcePull,
	}

	return c, nil
}

func (c *Case) Exec(ctx context.Context) error {
	var dockerImages []string
	if common.ArchArm64 == "arm64" {
		dockerImages = armRequiredImages
	} else {
		dockerImages = requiredImages
	}
	for _, img := range dockerImages {
		if !c.forcePull {
			if err := c.mng.CheckImageExists(ctx, img); err == nil {
				log.Printf("image %s found. skip\n", img)
				continue
			} else {
				if !errors.Is(err, common.ErrNotFound) {
					return err
				}
			}
		}

		log.Printf("pull image %s\n", img)
		if err := c.mng.PullImage(ctx, img); err != nil {
			log.Println(err)
		}
	}

	return nil
}
