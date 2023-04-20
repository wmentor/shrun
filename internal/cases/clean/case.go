package clean

import (
	"context"
	"errors"

	"github.com/wmentor/shrun/internal/common"
)

var (
	ErrInvalidManager = errors.New("invalid image manager")
)

type Case struct {
	mng   ImageRemover
	all   bool
	force bool
}

func NewCase(mng ImageRemover) (*Case, error) {
	if mng == nil || mng == ImageRemover(nil) {
		return nil, ErrInvalidManager
	}
	return &Case{
		mng: mng,
	}, nil
}

func (c *Case) WithAllImages(all bool) *Case {
	c.all = all
	return c
}

func (c *Case) WithForce(force bool) *Case {
	c.force = force
	return c
}

func (c *Case) Exec(ctx context.Context) error {
	if err := c.mng.RemoveImage(ctx, common.GetNodeContainerName(), c.force); err != nil {
		return err
	}
	return nil
}
