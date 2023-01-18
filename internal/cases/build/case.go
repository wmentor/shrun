package build

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/wmentor/shrun/internal/common"
)

var (
	ErrInvalidImageBuilder = errors.New("invalid image manager")
)

type Case struct {
	buildBasic bool
	buildPG    bool
	builder    ImageBuilder
}

func NewCase(builder ImageBuilder) (*Case, error) {
	if builder == nil || builder == ImageBuilder(nil) {
		return nil, ErrInvalidImageBuilder
	}

	return &Case{builder: builder}, nil
}

func (c *Case) WithBuildBasic() *Case {
	c.buildBasic = true
	return c.WithBuildPG()
}

func (c *Case) WithBuildPG() *Case {
	c.buildPG = true
	return c
}

func (c *Case) Exec(ctx context.Context) error {
	for _, copyFile := range []string{common.SpecFile, common.RcLocalFile} {
		if err := common.CopyFile(ctx, filepath.Join(common.GetConfigDir(), copyFile), filepath.Join(common.GetDataDir(), copyFile)); err != nil {
			return err
		}
	}

	if c.buildBasic {
		if err := c.builder.BuildImage(ctx, common.DockerfileGoBuilder, "gobuilder:latest"); err != nil {
			return err
		}

		if err := c.builder.BuildImage(ctx, common.DockerfilePgBuildEnv, "pgbuildenv:latest"); err != nil {
			return err
		}

		if err := c.builder.BuildImage(ctx, common.DockerfilePgDestEnv, "pgdestenv:latest"); err != nil {
			return err
		}

		if err := c.builder.BuildImage(ctx, common.DockerfileEtcd, "etcd:latest"); err != nil {
			return err
		}
	}

	if c.buildPG {
		if err := c.builder.BuildImage(ctx, common.DockerfileSdmNode, "sdmnode:latest"); err != nil {
			return err
		}
	}

	if err := c.builder.BuildImage(ctx, common.DockerfileShardman, "shardman:latest"); err != nil {
		return err
	}

	return nil
}
