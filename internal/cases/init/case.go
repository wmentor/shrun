package init

import (
	"context"
	"errors"

	"github.com/wmentor/shrun/internal/entities"
)

var (
	ErrInvalidExporter = errors.New("invalid exporter")
)

type Case struct {
	e    Exporter
	opts entities.ExportFileSettings
}

func NewCase(e Exporter, opts entities.ExportFileSettings) (*Case, error) {
	if e == nil || e == Exporter(nil) {
		return nil, ErrInvalidExporter
	}

	c := &Case{
		e:    e,
		opts: opts,
	}

	return c, nil
}

func (c *Case) Exec(_ context.Context) error {
	return c.e.ExportFiles(c.opts)
}
