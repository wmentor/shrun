package stop

import (
	"context"
)

type NetworkManager interface {
	CheckNetworkExists(ctx context.Context) (bool, error)
	DeleteNetwork(ctx context.Context) error
}
