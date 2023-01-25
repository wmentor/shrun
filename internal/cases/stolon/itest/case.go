package itest

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/entities"
)

var (
	ErrInvalidImageBuilder     = errors.New("invalid image builder")
	ErrInvalidContainerManager = errors.New("invalid container manager")
	ErrInvalidStatusCode       = errors.New("invalid status code")
)

type Case struct {
	builder  ImageBuilder
	cmanager ContainerManager
	tests    []string
}

func NewCase(builder ImageBuilder, containerManager ContainerManager) (*Case, error) {
	if builder == nil || builder == ImageBuilder(nil) {
		return nil, ErrInvalidImageBuilder
	}

	if containerManager == nil || containerManager == ContainerManager(nil) {
		return nil, ErrInvalidContainerManager
	}

	myCase := &Case{
		builder:  builder,
		cmanager: containerManager,
	}

	return myCase, nil
}

func (c *Case) WithTests(testNames []string) *Case {
	c.tests = testNames
	return c
}

func (c *Case) Exec(ctx context.Context) error {
	c.cmanager.StopAll(ctx)
	//defer c.cmanager.StopAll(ctx)

	if err := c.builder.BuildImage(ctx, common.DockerfileStolonInt, c.calcImageName()); err != nil {
		return fmt.Errorf("build stolon integration test image error: %w", err)
	}

	baseDir := "/build/shardman-utils"

	opts := entities.ContainerStartSettings{
		Image: c.calcImageName(),
		Host:  c.calcContainerName(),
		Envs: []string{
			"GOCACHE=/tmp/.cache/go-build",
			"GOLANGCI_LINT_CACHE=/tmp/.cache/golangci-lint",
			"STOLON_TEST_STORE_BACKEND=etcdv3",
			"ETCD_BIN=/opt/pgpro/sdm-14/bin/etcd",
			"BASEDIR=" + baseDir,
			"BINDIR=" + baseDir + "/bin",
			"STKEEPER_BIN=" + baseDir + "/bin/stolon-keeper",
			"STPROXY_BIN=" + baseDir + "/bin/stolon-proxy",
			"STSENTINEL_BIN=" + baseDir + "/bin/stolon-sentinel",
			"STCTL_BIN=" + baseDir + "/bin/stolonctl",
		},
	}

	cID, err := c.cmanager.CreateAndStart(ctx, opts)
	if err != nil {
		return fmt.Errorf("start container error: %w", err)
	}

	maker := strings.Builder{}

	if len(c.tests) > 0 {
		maker.WriteString("cd " + baseDir)
		maker.WriteString(" ; ")
		maker.WriteString("go test -timeout 60m  -v -count 1 -parallel 1")
		maker.WriteString(" ./tests/integration/utils.go")
		for _, tn := range c.tests {
			maker.WriteString(" ./tests/integration/")
			maker.WriteString(tn)
			maker.WriteString(".go")
		}
	} else {
		maker.WriteString(baseDir + "/tests/run_integration")
	}

	code, err := c.cmanager.Exec(ctx, cID, maker.String(), "root")
	if err != nil {
		return fmt.Errorf("stolon integration test failed: %w", err)
	}

	if code != 0 {
		return ErrInvalidStatusCode
	}

	return nil
}

func (c *Case) calcImageName() string {
	return "stolonint:latest"
}

func (c *Case) calcContainerName() string {
	return common.GetObjectPrefix() + "stolonint"
}
