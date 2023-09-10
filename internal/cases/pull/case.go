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
	requiredImages = []string{
		"postgres:14",
		"ubuntu:" + common.UbuntuVersion,
		"golang:" + common.GoVersion,
		"quay.io/coreos/etcd:v" + common.EtcdVersion,
		"prom/prometheus",
		"grafana/grafana",
		"prometheuscommunity/postgres-exporter",
		"gcr.io/cadvisor/cadvisor:v0.47.2",
	}
)

var (
	armRequiredImages = []string{
		"arm64v8/postgres:14",
		"arm64v8/ubuntu:" + common.UbuntuVersion,
		"golang:" + common.GoVersion,
		"quay.io/coreos/etcd:v" + common.EtcdVersion + "-arm64",
		"prom/prometheus",
		"grafana/grafana",
		"prometheuscommunity/postgres-exporter",
		"gcr.io/cadvisor/cadvisor:v0.47.2",
	}
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
	if common.WorkArch == common.ArchArm64 {
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
