package build

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/wmentor/shrun/internal/common"
)

var (
	ErrInvalidImageBuilder = errors.New("invalid image manager")
)

type Case struct {
	buildBasic bool
	buildPG    bool
	buildGoTpc bool
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

func (c *Case) WithGoTpc() *Case {
	c.buildGoTpc = true
	return c
}

func (c *Case) Exec(ctx context.Context) error {
	files := []string{common.SpecFile, common.RcLocalFile, common.OpenSSLConf, common.PrometheusConf, common.GrafanaDatasourceConf}

	for _, copyFile := range files {
		dest := filepath.Join(common.GetDataDir(), copyFile)
		os.Remove(dest)
		if err := common.CopyFile(ctx, filepath.Join(common.GetConfigDir(), copyFile), dest); err != nil {
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

		if err := c.builder.BuildImage(ctx, common.DockerfilePrometheus, "prometheus:latest"); err != nil {
			return err
		}

		if err := c.builder.BuildImage(ctx, common.DockerfileGrafana, "grafana:latest"); err != nil {
			return err
		}
	}

	if c.buildGoTpc || c.buildBasic {
		if err := c.builder.BuildImage(ctx, common.DockerfileGoTpc, "gotpc:latest"); err != nil {
			return err
		}
	}

	if c.buildPG {
		if err := c.builder.BuildImage(ctx, common.DockerfileSdmNode, common.GetSdmNodeImageName()); err != nil {
			return err
		}
	}

	if err := c.builder.BuildImage(ctx, common.DockerfileShardman, common.GetNodeContainerName()); err != nil {
		return err
	}

	for _, f := range files {
		os.Remove(filepath.Join(common.GetDataDir(), f))
	}

	return nil
}
