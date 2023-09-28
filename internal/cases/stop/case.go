package stop

import (
	"context"
	"errors"
)

var (
	ErrInvalidContainerManager = errors.New("invalid container manager")
	ErrInvalidNetworkManager   = errors.New("invalid network manager")
)

type Case struct {
	cmng   ContainerManager
	nmng   NetworkManager
	remove bool
}

func NewCase(cmng ContainerManager, nmng NetworkManager) (*Case, error) {
	if cmng == nil || cmng == ContainerManager(nil) {
		return nil, ErrInvalidContainerManager
	}

	if nmng == nil || nmng == NetworkManager(nil) {
		return nil, ErrInvalidNetworkManager
	}

	return &Case{
		nmng:   nmng,
		cmng:   cmng,
		remove: true,
	}, nil
}

func (c *Case) WithRemove(need bool) *Case {
	c.remove = need
	return c
}

func (c *Case) Exec(ctx context.Context) error {
	c.cmng.StopAll(ctx, c.remove)

	if c.remove {
		started, err := c.nmng.CheckNetworkExists(ctx)
		if err != nil {
			return err
		}

		if started {
			return c.nmng.DeleteNetwork(ctx)
		}
	}

	return nil
}
